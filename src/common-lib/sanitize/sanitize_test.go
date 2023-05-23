package sanitize

import (
	"testing"
)

func TestIdentifier(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "1", args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}, want: ""},
		{name: "2", args: args{s: "𝕿𝖍𝖊 𝖖𝖚𝖎𝖈𝖐 𝖇𝖗𝖔𝖜𝖓 𝖋𝖔𝖝 𝖏𝖚𝖒𝖕𝖘 𝖔𝖛𝖊𝖗 𝖙𝖍𝖊 𝖑𝖆𝖟𝖞 𝖉𝖔𝖌		"}, want: ""},
		{name: "3", args: args{s: "𝑻𝒉𝒆 𝒒𝒖𝒊𝒄𝒌 𝒃𝒓𝒐𝒘𝒏 𝒇𝒐𝒙 𝒋𝒖𝒎𝒑𝒔 𝒐𝒗𝒆𝒓 𝒕𝒉𝒆 𝒍𝒂𝒛𝒚 𝒅𝒐𝒈	"}, want: ""},
		{name: "4", args: args{s: "𝓣𝓱𝓮 𝓺𝓾𝓲𝓬𝓴 𝓫𝓻𝓸𝔀𝓷 𝓯𝓸𝔁 𝓳𝓾𝓶𝓹𝓼 𝓸𝓿𝓮𝓻 𝓽𝓱𝓮 𝓵𝓪𝔃𝔂 𝓭𝓸𝓰		"}, want: ""},
		{name: "5", args: args{s: "𝕋𝕙𝕖 𝕢𝕦𝕚𝕔𝕜 𝕓𝕣𝕠𝕨𝕟 𝕗𝕠𝕩 𝕛𝕦𝕞𝕡𝕤 𝕠𝕧𝕖𝕣 𝕥𝕙𝕖 𝕝𝕒𝕫𝕪 𝕕𝕠𝕘		"}, want: ""},
		{name: "6", args: args{s: "𝚃𝚑𝚎 𝚚𝚞𝚒𝚌𝚔 𝚋𝚛𝚘𝚠𝚗 𝚏𝚘𝚡 𝚓𝚞𝚖𝚙𝚜 𝚘𝚟𝚎𝚛 𝚝𝚑𝚎 𝚕𝚊𝚣𝚢 𝚍𝚘𝚐		"}, want: ""},
		{name: "7", args: args{s: "⒯⒣⒠ ⒬⒰⒤⒞⒦ ⒝⒭⒪⒲⒩ ⒡⒪⒳ ⒥⒰⒨⒫⒮ ⒪⒱⒠⒭ ⒯⒣⒠ ⒧⒜⒵⒴ ⒟⒪⒢		"}, want: ""},
		{name: "8", args: args{s: "<script>alert(123)</script>		"}, want: "script_alert_123_script"},
		{name: "9", args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;		"}, want: "lt_script_gt_alert_39_123_39_lt_script_gt"},
		{name: "10", args: args{s: "test"}, want: "test"},
		{name: "11", args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"}, want: ""},
		{name: "12", args: args{s: "--"}, want: ""},
		{name: "13", args: args{s: "123456789012345678901234567890123456789"}, want: "_123456789012345678901234567890123456789"},
		{name: "14", args: args{s: "--version"}, want: "version"},
		{name: "15", args: args{s: "$USER"}, want: "USER"},

		// Currect Strings
		{name: "16", args: args{s: "version"}, want: "version"},
		{name: "17", args: args{s: "TestMachine_juno:example_c counter"}, want: "TestMachine_juno_example_c_counter"},
		{name: "18", args: args{s: "platform-publisher-service-daemon-container-5fbc5c79cd-2l6gf"}, want: "platform_publisher_service_daemon_container_5fbc5c79cd_2l6gf"},
		{name: "19", args: args{s: "platform-publisher-service_v2"}, want: "platform_publisher_service_v2"},
		{name: "20", args: args{s: "Test"}, want: "Test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Identifier(tt.args.s); got != tt.want {
				t.Errorf("Identifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTML(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "1", args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}, want: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"},
		{name: "2", args: args{s: "<script>alert(123)</script>"}, want: "alert(123)"},
		{name: "3", args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}, want: "&lt;script&gt;alert('123');&lt;/script&gt;"},
		{name: "4", args: args{s: "ABC<div style=\"x:\xE2\x80\x8Bexpression(javascript:alert(1)\">DEF"}, want: "ABCDEF"},
		{name: "5", args: args{s: "<a href=\"javascript\x00:javascript:alert(1)\" id=\"fuzzelement1\">test</a>"}, want: "test"},

		// Currect Strings
		{name: "6", args: args{s: "test"}, want: "test"},
		{name: "7", args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"}, want: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"},
		{name: "8", args: args{s: "--"}, want: "--"},
		{name: "9", args: args{s: "1234567890123-45678901234567890123456789"}, want: "1234567890123-45678901234567890123456789"},
		{name: "10", args: args{s: "--version"}, want: "--version"},
		{name: "11", args: args{s: "$USER"}, want: "$USER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTML(tt.args.s); got != tt.want {
				t.Errorf("HTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTMLAllowing(t *testing.T) {
	type args struct {
		s    string
		args [][]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Naughty Strings
		{name: "1", args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}, want: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"},
		{name: "2", args: args{s: "<script>alert(123)</script>"}, want: ""},
		{name: "3", args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}, want: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"},
		{name: "4", args: args{s: "ABC<div style=\"x:\xE2\x80\x8Bexpression(javascript:alert(1)\">DEF"}, want: "ABC<div>DEF"},
		{name: "5", args: args{s: "<a href=\"javascript\x00:javascript:alert(1)\" id=\"fuzzelement1\">test</a>"}, want: "<a id=\"fuzzelement1\">test</a>"},
		{name: "2", args: args{s: "<i>hello world</i href=\"javascript:alert('hello world')\">"}, want: "<i>hello world</i>"},

		// Currect Strings
		{name: "6", args: args{s: "test"}, want: "test"},
		{name: "7", args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"}, want: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"},
		{name: "8", args: args{s: "--"}, want: "--"},
		{name: "9", args: args{s: "1234567890123-45678901234567890123456789"}, want: "1234567890123-45678901234567890123456789"},
		{name: "10", args: args{s: "--version"}, want: "--version"},
		{name: "11", args: args{s: "$USER"}, want: "$USER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLAllowing(tt.args.s, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTMLAllowing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HTMLAllowing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "1", args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}, want: "-"},
		{name: "2", args: args{s: "<script>alert(123)</script>"}, want: "scriptalert123-script"},
		{name: "3", args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}, want: "-ltscript-gtalert-39123-39-lt-script-gt"},
		{name: "4", args: args{s: "ABC<div style=\"x:\xE2\x80\x8Bexpression(javascript:alert(1)\">DEF"}, want: "ABCdiv-style-x-expressionjavascript-alert1DEF"},
		{name: "5", args: args{s: "<a href=\"javascript\x00:javascript:alert(1)\" id=\"fuzzelement1\">test</a>"}, want: "a-href-javascript-javascript-alert1-id-fuzzelement1test-a"},
		{name: "2", args: args{s: "<i>hello world</i href=\"javascript:alert('hello world')\">"}, want: "ihello-world-i-href-javascript-alerthello-world"},

		// Currect Strings
		{name: "6", args: args{s: "test"}, want: "test"},
		{name: "7", args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"}, want: "-"},
		{name: "8", args: args{s: "--"}, want: "-"},
		{name: "9", args: args{s: "1234567890123-45678901234567890123456789"}, want: "1234567890123-45678901234567890123456789"},
		{name: "10", args: args{s: "--version"}, want: "-version"},
		{name: "11", args: args{s: "$USER"}, want: "USER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Name(tt.args.s); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "1", args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}, want: "-"},
		{name: "2", args: args{s: "<script>alert(123)</script>"}, want: "script"},
		{name: "3", args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}, want: "script-gt"},
		{name: "4", args: args{s: "ABC<div style=\"x:\xE2\x80\x8Bexpression(javascript:alert(1)\">DEF"}, want: "abcdiv-style-x-expressionjavascript-alert1def"},
		{name: "5", args: args{s: "<a href=\"javascript\x00:javascript:alert(1)\" id=\"fuzzelement1\">test</a>"}, want: "a"},
		{name: "2", args: args{s: "<i>hello world</i href=\"javascript:alert('hello world')\">"}, want: "i-href-javascript-alerthello-world"},

		// Currect Strings
		{name: "6", args: args{s: "test"}, want: "test"},
		{name: "7", args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِيَّةُ،"}, want: "-"},
		{name: "8", args: args{s: "--"}, want: "-"},
		{name: "9", args: args{s: "1234567890123-45678901234567890123456789"}, want: "1234567890123-45678901234567890123456789"},
		{name: "10", args: args{s: "--version"}, want: "-version"},
		{name: "11", args: args{s: "$USER"}, want: "user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileName(tt.args.s); got != tt.want {
				t.Errorf("FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "1", args: args{s: "ReAd ME.md"}, want: `read-me.md`},
		{name: "2", args: args{s: "E88E08A7-279C-4CC1-8B90-86DE0D7044_3C.html"}, want: `e88e08a7-279c-4cc1-8b90-86de0d7044-3c.html`},
		{name: "3", args: args{s: "/user/test/I am a long url's_-?ASDF@£$%£%^testé.html"}, want: `/user/test/i-am-a-long-urls-asdfteste.html`},
		{name: "4", args: args{s: "/../../4-icon.jpg"}, want: `/4-icon.jpg`},
		{name: "5", args: args{s: "/Images_dir/../4-icon.jpg"}, want: `/images-dir/4-icon.jpg`},
		{name: "6", args: args{s: "../4 icon.*"}, want: `/4-icon.`},
		{name: "7", args: args{s: "Spac ey/Nôm/test før url"}, want: `spac-ey/nom/test-foer-url`},
		{name: "8", args: args{s: "../*"}, want: `/`},

		// Currect Strings
		{name: "9", args: args{s: "/test"}, want: "/test"},
		{name: "10", args: args{s: "/--version"}, want: "/-version"},
		{name: "11", args: args{s: "/USER"}, want: "/user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Path(tt.args.s); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumber(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Naughty Strings
		{name: "01", args: args{s: "ReAd ME.md"}, want: ``},
		{name: "02", args: args{s: "E88E08A7-279C-4CC1-8.B90-86.DE0D7044_3C.html"}, want: `88087279418.9086`},
		{name: "03", args: args{s: "/user/test/I am a long url's_-?ASDF@£$%£%^testé.html"}, want: ``},
		{name: "04", args: args{s: "/../..qw4-icon.jpg"}, want: `4`},
		{name: "05", args: args{s: "/Images_dir/..qw/234-icon.jpg"}, want: `234`},
		{name: "06", args: args{s: "testabcdbbk"}, want: ``},
		{name: "07", args: args{s: "Spac ey/Nôm/test før url"}, want: ``},

		// Currect Strings
		{name: "08", args: args{s: "0.23"}, want: `0.23`},
		{name: "09", args: args{s: "1234"}, want: "1234"},
		{name: "10", args: args{s: "1234.256"}, want: "1234.256"},
		{name: "11", args: args{s: "123456987.33"}, want: "123456987.33"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Number(tt.args.s); got != tt.want {
				t.Errorf("Number() = %v, want %v", got, tt.want)
			}
		})
	}
}
