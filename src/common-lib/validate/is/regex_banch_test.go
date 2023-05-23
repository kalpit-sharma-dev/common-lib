package is

import (
	"testing"
)

func BenchmarkPhone(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: ""}},
		{name: "02", want: false, args: args{s: "abc"}},
		{name: "03", want: false, args: args{s: "123-456-7890"}},
		{name: "04", want: false, args: args{s: "123-XXX-XXXX"}},
		{name: "05", want: false, args: args{s: "XXX-123-XXXX"}},
		{name: "06", want: false, args: args{s: "XXX-XXX-1123"}},
		{name: "07", want: false, args: args{s: "123-256-23658"}},
		{name: "08", want: false, args: args{s: "1234-123-125"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "+4974339296"}},
		{name: "10", want: true, args: args{s: "+1 (123) 456-7890"}},
		{name: "11", want: true, args: args{s: "0591 74339296"}},
		{name: "12", want: true, args: args{s: "+(591) (4) 6434850"}},
		{name: "13", want: true, args: args{s: "0001 5555555555"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := Phone(tt.args.s); got != tt.want {
				b.Errorf("Phone() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkEmail(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: ""}},
		{name: "02", want: false, args: args{s: "abc"}},
		{name: "03", want: false, args: args{s: "@com"}},
		{name: "04", want: false, args: args{s: "abc@abc"}},
		{name: "05", want: false, args: args{s: "abc@abccom"}},
		{name: "06", want: false, args: args{s: "abc.com"}},
		{name: "07", want: false, args: args{s: "abc@.com"}},
		{name: "08", want: false, args: args{s: "@abc.com"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "a@a.com"}},
		{name: "10", want: true, args: args{s: "abc@abc.com"}},
		{name: "11", want: true, args: args{s: "test@test.co.in"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := Email(tt.args.s); got != tt.want {
				b.Errorf("Email() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkUUID(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: ""}},
		{name: "02", want: false, args: args{s: "a987fbc9-4bed-3078-cf07"}},
		{name: "03", want: false, args: args{s: "4bed-3078-cf07-9141ba07c9f1"}},
		{name: "04", want: false, args: args{s: "a987fbc94bed-3078-cf07-9141ba07c9f1"}},
		{name: "05", want: false, args: args{s: "b987fbc9-4bed-3078-cf079141ba07c9f3"}},
		{name: "06", want: false, args: args{s: "57b73598-8764-4ad0-a76a-679bb6640e"}},
		{name: "07", want: false, args: args{s: "a987fbc9-4bed-3078cf07-9141ba07c9f1"}},
		{name: "08", want: false, args: args{s: "987fbc97-4bed5078-af07-9141ba07c9f3"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "a987fbc9-4bed-3078-cf07-9141ba07c9f1"}},
		{name: "10", want: true, args: args{s: "a987fbc9-4bed-3078-cf07-9141ba07c9f3"}},
		{name: "11", want: true, args: args{s: "987fbc97-4bed-5078-af07-9141ba07c9f3"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := UUID(tt.args.s); got != tt.want {
				b.Errorf("UUID() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkAlpha(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: ""}},
		{name: "02", want: false, args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}},
		{name: "03", want: false, args: args{s: "⒯⒣⒠ ⒬⒰⒤⒞⒦ ⒝⒭⒪⒲⒩ ⒡⒪⒳ ⒥⒰⒨⒫⒮ ⒪⒱⒠⒭ ⒯⒣⒠ ⒧⒜⒵⒴ ⒟⒪⒢"}},
		{name: "04", want: false, args: args{s: "<script>alert(123)</script>"}},
		{name: "05", want: false, args: args{s: "𝕋𝕙𝕖 𝕢𝕦𝕚𝕔𝕜 𝕓𝕣𝕠𝕨𝕟 𝕗𝕠𝕩 𝕛𝕦𝕞𝕡𝕤 𝕠𝕧𝕖𝕣 𝕥𝕙𝕖 𝕝𝕒𝕫𝕪 𝕕𝕠𝕘"}},
		{name: "06", want: false, args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}},
		{name: "07", want: false, args: args{s: "123456789012345678901234567890123456789"}},
		{name: "08", want: false, args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِ"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "version"}},
		{name: "10", want: true, args: args{s: "Test"}},
		{name: "11", want: true, args: args{s: "platformpublisherservice"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := Alpha(tt.args.s); got != tt.want {
				b.Errorf("Alpha() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkAlphaNumeric(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: ""}},
		{name: "02", want: false, args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}},
		{name: "03", want: false, args: args{s: "⒯⒣⒠ ⒬⒰⒤⒞⒦ ⒝⒭⒪⒲⒩ ⒡⒪⒳ ⒥⒰⒨⒫⒮ ⒪⒱⒠⒭ ⒯⒣⒠ ⒧⒜⒵⒴ ⒟⒪⒢"}},
		{name: "04", want: false, args: args{s: "<script>alert(123)</script>"}},
		{name: "05", want: false, args: args{s: "𝕋𝕙𝕖 𝕢𝕦𝕚𝕔𝕜 𝕓𝕣𝕠𝕨𝕟 𝕗𝕠𝕩 𝕛𝕦𝕞𝕡𝕤 𝕠𝕧𝕖𝕣 𝕥𝕙𝕖 𝕝𝕒𝕫𝕪 𝕕𝕠𝕘"}},
		{name: "06", want: false, args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}},
		{name: "07", want: false, args: args{s: "12345678901234567_8901234567890123456789"}},
		{name: "08", want: false, args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِ"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "version121"}},
		{name: "10", want: true, args: args{s: "Test"}},
		{name: "11", want: true, args: args{s: "123456789012345678901234567890123456789"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := AlphaNumeric(tt.args.s); got != tt.want {
				b.Errorf("AlphaNumeric() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkNumber(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: "123456987.33.333"}},
		{name: "02", want: false, args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}},
		{name: "03", want: false, args: args{s: "⒯⒣⒠ ⒬⒰⒤⒞⒦ ⒝⒭⒪⒲⒩ ⒡⒪⒳ ⒥⒰⒨⒫⒮ ⒪⒱⒠⒭ ⒯⒣⒠ ⒧⒜⒵⒴ ⒟⒪⒢"}},
		{name: "04", want: false, args: args{s: "<script>alert(123)</script>"}},
		{name: "05", want: false, args: args{s: "𝕋𝕙𝕖 𝕢𝕦𝕚𝕔𝕜 𝕓𝕣𝕠𝕨𝕟 𝕗𝕠𝕩 𝕛𝕦𝕞𝕡𝕤 𝕠𝕧𝕖𝕣 𝕥𝕙𝕖 𝕝𝕒𝕫𝕪 𝕕𝕠𝕘"}},
		{name: "06", want: false, args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}},
		{name: "07", want: false, args: args{s: "12345678901234567_8901234567890123456789"}},
		{name: "08", want: false, args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِ"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "88087279418.9086"}},
		{name: "10", want: true, args: args{s: "123456987.33"}},
		{name: "11", want: true, args: args{s: "123456789012345678901234567890123456789"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := Number(tt.args.s); got != tt.want {
				b.Errorf("Number() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}

func BenchmarkIdentifier(b *testing.B) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Naughty Strings
		{name: "01", want: false, args: args{s: "123456987.33.333"}},
		{name: "02", want: false, args: args{s: "𝐓𝐡𝐞 𝐪𝐮𝐢𝐜𝐤 𝐛𝐫𝐨𝐰𝐧 𝐟𝐨𝐱 𝐣𝐮𝐦𝐩𝐬 𝐨𝐯𝐞𝐫 𝐭𝐡𝐞 𝐥𝐚𝐳𝐲 𝐝𝐨𝐠"}},
		{name: "03", want: false, args: args{s: "⒯⒣⒠ ⒬⒰⒤⒞⒦ ⒝⒭⒪⒲⒩ ⒡⒪⒳ ⒥⒰⒨⒫⒮ ⒪⒱⒠⒭ ⒯⒣⒠ ⒧⒜⒵⒴ ⒟⒪⒢"}},
		{name: "04", want: false, args: args{s: "<script>alert(123)</script>"}},
		{name: "05", want: false, args: args{s: "𝕋𝕙𝕖 𝕢𝕦𝕚𝕔𝕜 𝕓𝕣𝕠𝕨𝕟 𝕗𝕠𝕩 𝕛𝕦𝕞𝕡𝕤 𝕠𝕧𝕖𝕣 𝕥𝕙𝕖 𝕝𝕒𝕫𝕪 𝕕𝕠𝕘"}},
		{name: "06", want: false, args: args{s: "&lt;script&gt;alert(&#39;123&#39;);&lt;/script&gt;"}},
		{name: "07", want: false, args: args{s: "12345678901234567_8901234567890123456789"}},
		{name: "08", want: false, args: args{s: "مُنَاقَشَةُ سُبُلِ اِسْتِخْدَامِ اللُّغَةِ فِي النُّظُمِ الْقَائِمَةِ وَفِيم يَخُصَّ التَّطْبِيقَاتُ الْحاسُوبِ"}},

		// Currect Strings
		{name: "09", want: true, args: args{s: "lt_script_gt_alert_39_123_39_lt_script_gt"}},
		{name: "10", want: true, args: args{s: "version"}},
		{name: "11", want: true, args: args{s: "_123456789012345678901234567890123456789"}},
	}
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			if got := Identifier(tt.args.s); got != tt.want {
				b.Errorf("Identifier() %s = %v, want %v", tt.name, got, tt.want)
			}
		}
	}
}
