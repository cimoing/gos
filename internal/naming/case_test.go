package naming

import "testing"

func TestCaseConversions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		snake  string
		kebab  string
		pascal string
		camel  string
	}{
		{
			name:   "snake input",
			input:  "user_profile",
			snake:  "user_profile",
			kebab:  "user-profile",
			pascal: "UserProfile",
			camel:  "userProfile",
		},
		{
			name:   "kebab input",
			input:  "user-profile",
			snake:  "user_profile",
			kebab:  "user-profile",
			pascal: "UserProfile",
			camel:  "userProfile",
		},
		{
			name:   "pascal input",
			input:  "UserProfile",
			snake:  "user_profile",
			kebab:  "user-profile",
			pascal: "UserProfile",
			camel:  "userProfile",
		},
		{
			name:   "path input",
			input:  "user/register",
			snake:  "user_register",
			kebab:  "user-register",
			pascal: "UserRegister",
			camel:  "userRegister",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnake(tt.input); got != tt.snake {
				t.Fatalf("ToSnake() = %q, want %q", got, tt.snake)
			}
			if got := ToKebab(tt.input); got != tt.kebab {
				t.Fatalf("ToKebab() = %q, want %q", got, tt.kebab)
			}
			if got := ToPascal(tt.input); got != tt.pascal {
				t.Fatalf("ToPascal() = %q, want %q", got, tt.pascal)
			}
			if got := ToCamel(tt.input); got != tt.camel {
				t.Fatalf("ToCamel() = %q, want %q", got, tt.camel)
			}
		})
	}
}
