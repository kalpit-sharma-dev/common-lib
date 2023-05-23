package token

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient/mock"
)

func TestDefaultJWTExchanger_Exchange(t *testing.T) {
	var (
		pid = "pid"
		jwt = "some.jwt.token"
	)

	t.Run("positive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		wClient := mock.NewMockClientService(ctrl)

		response, err := json.Marshal(Token{Value: jwt})
		require.NoError(t, err)
		body := ioutil.NopCloser(bytes.NewReader(response))

		wClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}, nil)

		cfg := AuthorizationConfig{
			URL:    "http://127.0.0.1",
			Client: wClient,
		}
		target := NewJWTExchanger(cfg, nil)

		jwtToken, err := target.Exchange(context.Background(), "", pid)
		require.NoError(t, err)
		require.Equal(t, jwt, jwtToken, "got invalid token")
	})

	t.Run("negative_invalid_response", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		cfg := AuthorizationConfig{
			URL:    "http://127.0.0.1",
			Client: http.DefaultClient,
		}

		target := NewJWTExchanger(cfg, nil)

		httpmock.RegisterResponder(http.MethodGet,
			fmt.Sprintf(exchangeJWTFmtURL, cfg.URL, pid),
			httpmock.NewBytesResponder(http.StatusInternalServerError, nil))

		_, err := target.Exchange(context.Background(), "", pid)
		require.Error(t, err)
	})
}

func TestJWTExchanger_ServiceTokenExchange(t *testing.T) {
	type fields struct {
		URL     string
		Client  func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService
		Retrier *retrier.Retrier
	}
	type args struct {
		ctx context.Context
		req AssumeRoleRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				URL: "http://127.0.0.1",
				Client: func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService {
					client := mock.NewMockClientService(ctrl)

					response, err := json.Marshal(Token{Value: "some.jwt.token"})
					require.NoError(t, err)
					body := ioutil.NopCloser(bytes.NewReader(response))

					client.EXPECT().Do(gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil)
					return client
				},
				Retrier: nil, // default retrier
			},
			args: args{
				ctx: context.Background(),
				req: AssumeRoleRequest{
					RoleName: "Admin",
				},
			},
			want:    "some.jwt.token",
			wantErr: false,
		},
		{
			name: "positive with user",
			fields: fields{
				URL: "http://127.0.0.1",
				Client: func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService {
					client := mock.NewMockClientService(ctrl)

					response, err := json.Marshal(Token{Value: "some.jwt.token"})
					require.NoError(t, err)
					body := ioutil.NopCloser(bytes.NewReader(response))

					client.EXPECT().Do(gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       body,
					}, nil)
					return client
				},
				Retrier: nil, // default retrier
			},
			args: args{
				ctx: context.Background(),
				req: AssumeRoleRequest{
					RoleName:  "Admin",
					UserID:    "admin@user.id",
					PartnerID: "admin-partner",
				},
			},
			want:    "some.jwt.token",
			wantErr: false,
		},
		{
			name: "failed to unmarshal response from Auth MS",
			fields: fields{
				URL: "http://127.0.0.1",
				Client: func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService {
					client := mock.NewMockClientService(ctrl)

					client.EXPECT().Do(gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{")), // malformed JSON response
					}, nil)
					return client
				},
				Retrier: nil,
			},
			args: args{
				ctx: context.Background(),
				req: AssumeRoleRequest{
					RoleName: "Admin",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "HTTP request and all retries are failed due to networking issue",
			fields: fields{
				URL: "http://127.0.0.1",
				Client: func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService {
					client := mock.NewMockClientService(ctrl)

					client.EXPECT().Do(gomock.Any()).Return(nil, &url.Error{
						Op:  "Get",
						URL: "http://127.0.0.1",
						Err: stderrors.New("boom"),
					}).Times(3) // 1 http call + 2 retries
					return client
				},
				Retrier: retrier.New(retrier.ConstantBackoff(2, 0), nil), // 2 retries with 0 delay
			},
			args: args{
				ctx: context.Background(),
				req: AssumeRoleRequest{
					RoleName: "Admin",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "HTTP request and retries are failed due to 500 HTTP response",
			fields: fields{
				URL: "http://127.0.0.1",
				Client: func(t *testing.T, ctrl *gomock.Controller) webClient.ClientService {
					client := mock.NewMockClientService(ctrl)

					client.EXPECT().Do(gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("500 HTTP error")),
					}, nil).Times(3) // 1 http call + 2 retries
					return client
				},
				Retrier: retrier.New(retrier.ConstantBackoff(2, 0), nil), // 2 retries with 0 delay
			},
			args: args{
				ctx: context.Background(),
				req: AssumeRoleRequest{
					RoleName: "Admin",
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := AuthorizationConfig{
				URL:     tt.fields.URL,
				Client:  tt.fields.Client(t, ctrl),
				Retrier: tt.fields.Retrier,
			}
			exchanger := NewJWTExchanger(cfg, nil)

			got, err := exchanger.AssumeRole(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestJWTExchanger_ExchangeEnhanced(t *testing.T) {
	var (
		pid     = "pid"
		jwt     = "some.jwt.token"
		mockURL = "http://127.0.0.1"
	)

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		wClient := mock.NewMockClientService(ctrl)

		response, err := json.Marshal(Token{Value: jwt})
		require.NoError(t, err)
		body := ioutil.NopCloser(bytes.NewReader(response))

		wClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
		}, nil)

		cfg := AuthorizationConfig{
			URL:    mockURL,
			Client: wClient,
		}
		target := NewJWTExchanger(cfg, logger.DiscardLogger())

		jwtToken, err := target.ExchangeEnhanced(context.Background(), "", pid)
		require.NoError(t, err)
		require.Equal(t, jwt, jwtToken, "invalid token received")
	})

	t.Run("Failure", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		cfg := AuthorizationConfig{
			URL:    mockURL,
			Client: http.DefaultClient,
		}

		target := NewJWTExchanger(cfg, logger.DiscardLogger())

		httpmock.RegisterResponder(http.MethodGet,
			fmt.Sprintf(exchangeJWTFmtURL, cfg.URL, pid),
			httpmock.NewBytesResponder(http.StatusInternalServerError, nil))

		_, err := target.ExchangeEnhanced(context.Background(), "", pid)
		require.Error(t, err)
	})
}
