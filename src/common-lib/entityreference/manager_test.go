package entityreference

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func TestNewManager(t *testing.T) {
	assert.NotNil(t, NewManagementUsecase(nil, nil, logger.DiscardLogger()))
}

func TestManager_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo Repo
	}
	type args struct {
		reference   *ReferenceRequest
		partnerID   string
		entityID    gocql.UUID
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().AddOne(&Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}).Return(nil)
					return mockRepo
				}(),
			},
			args: args{
				reference: &ReferenceRequest{
					ReferencingObjectID:     id,
					Type:                    Hard,
					ValidationCallbackURL:   "https://google.com",
					NotificationCallbackURL: "",
				},
				partnerID:   id.String(),
				entityID:    id,
				serviceName: "ears",
			},
			wantErr: false,
		},
		{
			name: "failed_insert",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().AddOne(&Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}).Return(errors.New("err"))
					return mockRepo
				}(),
			},
			args: args{
				reference: &ReferenceRequest{
					ReferencingObjectID:     id,
					Type:                    Hard,
					ValidationCallbackURL:   "https://google.com",
					NotificationCallbackURL: "",
				},
				partnerID:   id.String(),
				entityID:    id,
				serviceName: "ears",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, nil, logger.DiscardLogger())
			err := m.Create(context.Background(), tt.args.reference, tt.args.entityID, tt.args.serviceName, tt.args.partnerID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo Repo
	}
	type args struct {
		entityID    gocql.UUID
		serviceName string
		referenceID gocql.UUID
		partnerID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().DeleteOne(&Reference{
						PartnerID:           id.String(),
						EntityID:            id,
						Service:             "ears",
						ReferencingObjectID: id,
					}).Return(nil)
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				serviceName: "ears",
				referenceID: id,
				partnerID:   id.String(),
			},
			wantErr: false,
		},
		{
			name: "delete_failed",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().DeleteOne(&Reference{
						PartnerID:           id.String(),
						EntityID:            id,
						Service:             "ears",
						ReferencingObjectID: id,
					}).Return(errors.New("err"))
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				serviceName: "ears",
				referenceID: id,
				partnerID:   id.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, nil, logger.DiscardLogger())
			err := m.Delete(context.Background(), tt.args.entityID, tt.args.referenceID, tt.args.serviceName, tt.args.partnerID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo Repo
	}
	type args struct {
		entityID    gocql.UUID
		referenceID gocql.UUID
		serviceName string
		partnerID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ReferenceResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetOne(id.String(), id, "ears", id).Return(&Reference{
						EntityID:              id,
						Service:               "ears",
						ReferencingObjectID:   id,
						Type:                  Hard,
						ValidationCallbackURL: "https://google.com",
					}, nil)
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				referenceID: id,
				serviceName: "ears",
				partnerID:   id.String(),
			},
			want: &ReferenceResponse{
				EntityID:            id,
				ReferencingObjectID: id,
				Service:             "ears",
				Type:                Hard,
			},
			wantErr: false,
		},
		{
			name: "getByID_failed",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetOne(id.String(), id, "ears", id).Return(nil, errors.New("err"))
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				referenceID: id,
				serviceName: "ears",
				partnerID:   id.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, nil, logger.DiscardLogger())
			response, err := m.Get(context.Background(), tt.args.entityID, tt.args.referenceID, tt.args.serviceName, tt.args.partnerID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, response)
			}
		})
	}
}

func TestManager_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo Repo
	}
	type args struct {
		entityID    gocql.UUID
		serviceName string
		partnerID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*ReferenceResponse
		wantErr bool
	}{
		{
			name: "success_with_service_name",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetReferences(id.String(), id, "ears").Return([]*Reference{
						{
							PartnerID:             id.String(),
							EntityID:              id,
							Service:               "ears",
							ReferencingObjectID:   id,
							Type:                  Hard,
							ValidationCallbackURL: "https://google.com",
						},
					}, nil)
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				serviceName: "ears",
				partnerID:   id.String(),
			},
			want: []*ReferenceResponse{
				{
					EntityID:            id,
					ReferencingObjectID: id,
					Service:             "ears",
					Type:                Hard,
				},
			},
			wantErr: false,
		},
		{
			name: "success_without_service_name",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetReferences(id.String(), id).Return([]*Reference{
						{
							PartnerID:             id.String(),
							EntityID:              id,
							Service:               "ears",
							ReferencingObjectID:   id,
							Type:                  Hard,
							ValidationCallbackURL: "https://google.com",
						},
					}, nil)
					return mockRepo
				}(),
			},
			args: args{
				entityID:  id,
				partnerID: id.String(),
			},
			want: []*ReferenceResponse{
				{
					EntityID:            id,
					ReferencingObjectID: id,
					Service:             "ears",
					Type:                Hard,
				},
			},
			wantErr: false,
		},
		{
			name: "failed_with_service_name",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetReferences(id.String(), id, "ears").Return(nil, errors.New("err"))
					return mockRepo
				}(),
			},
			args: args{
				entityID:    id,
				serviceName: "ears",
				partnerID:   id.String(),
			},
			wantErr: true,
		},
		{
			name: "failed_with_service_name",
			fields: fields{
				Repo: func() Repo {
					mockRepo := NewMockRepo(ctrl)
					mockRepo.EXPECT().GetReferences(id.String(), id).Return(nil, errors.New("err"))
					return mockRepo
				}(),
			},
			args: args{
				entityID:  id,
				partnerID: id.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, nil, logger.DiscardLogger())
			all, err := m.GetAll(context.Background(), tt.args.entityID, tt.args.serviceName, tt.args.partnerID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, all)
			}
		})
	}
}

func TestManagementUsecase_CleanUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo   Repo
		Client Client
	}
	tests := []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "success_with_delete_reference",
			fields: fields{
				Repo: func() Repo {
					reference := &Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Soft,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{reference}, nil)
					repo.EXPECT().DeleteOne(reference).Return(nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			want: nil,
		},
		{
			name: "success_no_references",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					return client
				}(),
			},
			want: nil,
		},
		{
			name: "success_has_reference",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			want: nil,
		},
		{
			name: "failed_get_references",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return(nil, errors.New("err"))
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					return client
				}(),
			},
			want: errors.New("err"),
		},
		{
			name: "failed_make_request",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(nil, errors.New("err"))
					return client
				}(),
			},
			want: nil,
		},
		{
			name: "internal_error",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error":"internal","description":"desc"}`))),
					}, nil)
					return client
				}(),
			},
			want: nil,
		},
		{
			name: "failed_to_delete_reference",
			fields: fields{
				Repo: func() Repo {
					reference := &Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences().Return([]*Reference{reference}, nil)
					repo.EXPECT().DeleteOne(reference).Return(errors.New("err"))
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, tt.fields.Client, logger.DiscardLogger())
			err := m.CleanUp(context.Background())
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestManagementUsecase_CleanUpByPartner(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := gocql.TimeUUID()

	type fields struct {
		Repo   Repo
		Client Client
	}
	type args struct {
		partnerID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "success_with_delete_reference",
			fields: fields{
				Repo: func() Repo {
					reference := &Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Soft,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{reference}, nil)
					repo.EXPECT().DeleteOne(reference).Return(nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
		{
			name: "success_no_references",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
		{
			name: "success_has_reference",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
		{
			name: "failed_get_references",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return(nil, errors.New("err"))
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: errors.New("err"),
		},
		{
			name: "failed_get_partner_references",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(id.String()).Return(nil, errors.New("err"))
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					return client
				}(),
			},
			args: args{
				partnerID: id.String(),
			},
			want: errors.New("err"),
		},
		{
			name: "failed_make_request",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(nil, errors.New("err"))
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
		{
			name: "internal_error",
			fields: fields{
				Repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}}, nil)
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error":"internal","description":"desc"}`))),
					}, nil)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
		{
			name: "failed_to_delete_reference",
			fields: fields{
				Repo: func() Repo {
					reference := &Reference{
						PartnerID:               id.String(),
						EntityID:                id,
						Service:                 "ears",
						ReferencingObjectID:     id,
						Type:                    Hard,
						ValidationCallbackURL:   "https://google.com",
						NotificationCallbackURL: "",
					}
					repo := NewMockRepo(ctrl)
					repo.EXPECT().GetReferences(gomock.Any()).Return([]*Reference{reference}, nil)
					repo.EXPECT().DeleteOne(reference).Return(errors.New("err"))
					return repo
				}(),
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					}, nil)
					return client
				}(),
			},
			args: args{partnerID: id.String()},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.Repo, tt.fields.Client, logger.DiscardLogger())
			err := m.CleanUpByPartner(context.Background(), tt.args.partnerID)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestManagementUsecase_NotifySoftReference(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		Client Client
	}
	type args struct {
		reference *Reference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "success",
			fields: fields{
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
					}, nil)
					return client
				}(),
			},
			args: args{
				reference: &Reference{},
			},
			wantErr: nil,
		},
		{
			name: "failed",
			fields: fields{
				Client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.AssignableToTypeOf(&http.Request{})).Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error":"err","description":"err2"}`))),
					}, nil)
					return client
				}(),
			},
			args: args{
				reference: &Reference{},
			},
			wantErr: MsgError{
				Message: "request failed with status: 500",
				Desc:    `{"error":"err","description":"err2"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(nil, tt.fields.Client, logger.DiscardLogger())
			err := m.NotifySoftReference(context.Background(), tt.args.reference)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_getTransactionID(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, getTransactionID(ctx))
	ctx = context.WithValue(ctx, auth.TransactionKey, "some_value")
	assert.Equal(t, "some_value", getTransactionID(ctx))
}

func TestManagementUsecase_ValidateDeletion(t *testing.T) {
	ctrl := gomock.NewController(t)
	type fields struct {
		repo   Repo
		client Client
	}
	type args struct {
		references []*Reference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "success_soft_reference",
			fields: fields{
				repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().DeleteOne(gomock.Any()).Return(nil)
					return repo
				}(),
				client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
					}, nil)
					return client
				}(),
			},
			args: args{
				references: []*Reference{{
					PartnerID:               "",
					EntityID:                gocql.UUID{},
					Service:                 "",
					ReferencingObjectID:     gocql.UUID{},
					Type:                    Soft,
					ValidationCallbackURL:   "",
					NotificationCallbackURL: "https://google.com",
				}},
			},
			wantErr: nil,
		},
		{
			name: "success_with_notify_failed",
			fields: fields{
				repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().DeleteOne(gomock.Any()).Return(nil)
					return repo
				}(),
				client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
					}, nil)
					return client
				}(),
			},
			args: args{
				references: []*Reference{{
					PartnerID:               "",
					EntityID:                gocql.UUID{},
					Service:                 "",
					ReferencingObjectID:     gocql.UUID{},
					Type:                    Soft,
					ValidationCallbackURL:   "",
					NotificationCallbackURL: "https://google.com",
				}},
			},
			wantErr: nil,
		},
		{
			name: "delete_reference_failed",
			fields: fields{
				repo: func() Repo {
					repo := NewMockRepo(ctrl)
					repo.EXPECT().DeleteOne(gomock.Any()).Return(errors.New("err"))
					return repo
				}(),
				client: func() Client {
					client := NewMockClient(ctrl)
					client.EXPECT().Do(gomock.Any(), gomock.Any()).Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
					}, nil)
					return client
				}(),
			},
			args: args{
				references: []*Reference{{
					PartnerID:               "",
					EntityID:                gocql.UUID{},
					Service:                 "",
					ReferencingObjectID:     gocql.UUID{},
					Type:                    Soft,
					ValidationCallbackURL:   "",
					NotificationCallbackURL: "https://google.com",
				}},
			},
			wantErr: errors.New("err"),
		},
		{
			name: "conflict",
			fields: fields{
				repo:   nil,
				client: nil,
			},
			args: args{
				references: []*Reference{{
					PartnerID:               "",
					EntityID:                gocql.UUID{},
					Service:                 "",
					ReferencingObjectID:     gocql.UUID{},
					Type:                    Hard,
					ValidationCallbackURL:   "https://google.com",
					NotificationCallbackURL: "https://google.com",
				}},
			},
			wantErr: NewConflictReferenceError([]*Reference{{
				PartnerID:               "",
				EntityID:                gocql.UUID{},
				Service:                 "",
				ReferencingObjectID:     gocql.UUID{},
				Type:                    Hard,
				ValidationCallbackURL:   "https://google.com",
				NotificationCallbackURL: "https://google.com",
			}}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManagementUsecase(tt.fields.repo, tt.fields.client, logger.DiscardLogger())
			err := m.ValidateDeletion(context.Background(), tt.args.references)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
