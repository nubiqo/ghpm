package profile

import (
	"testing"
)

func TestProfile_Validate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: Profile{
				Name:          "test-profile",
				GitUsername:   "testuser",
				GitEmail:      "test@example.com",
				SSHPrivateKey: "private-key",
				SSHPublicKey:  "public-key",
			},
			wantErr: false,
		},
		{
			name: "missing ssh keys",
			profile: Profile{
				Name:        "test-profile",
				GitUsername: "testuser",
				GitEmail:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			profile: Profile{
				Name:        "",
				GitUsername: "testuser",
				GitEmail:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			profile: Profile{
				Name:        "test-profile",
				GitUsername: "",
				GitEmail:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty email",
			profile: Profile{
				Name:        "test-profile",
				GitUsername: "testuser",
				GitEmail:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			profile: Profile{
				Name:        "test-profile",
				GitUsername: "testuser",
				GitEmail:    "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "invalid characters in name",
			profile: Profile{
				Name:        "test/profile",
				GitUsername: "testuser",
				GitEmail:    "test@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Profile.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfile_HasSSHKeys(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		want    bool
	}{
		{
			name: "has both keys",
			profile: Profile{
				SSHPrivateKey: "private-key-content",
				SSHPublicKey:  "public-key-content",
			},
			want: true,
		},
		{
			name: "has only private key",
			profile: Profile{
				SSHPrivateKey: "private-key-content",
				SSHPublicKey:  "",
			},
			want: false,
		},
		{
			name: "has only public key",
			profile: Profile{
				SSHPrivateKey: "",
				SSHPublicKey:  "public-key-content",
			},
			want: false,
		},
		{
			name: "has no keys",
			profile: Profile{
				SSHPrivateKey: "",
				SSHPublicKey:  "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.HasSSHKeys(); got != tt.want {
				t.Errorf("Profile.HasSSHKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfile_Clone(t *testing.T) {
	original := &Profile{
		Name:          "original",
		GitUsername:   "testuser",
		GitEmail:      "test@example.com",
		SSHPrivateKey: "private-key",
		SSHPublicKey:  "public-key",
		IsActive:      true,
		CreatedFrom:   "manual",
	}

	cloned := original.Clone("cloned")

	if cloned.Name != "cloned" {
		t.Errorf("Clone() name = %v, want %v", cloned.Name, "cloned")
	}

	if cloned.GitUsername != original.GitUsername {
		t.Errorf("Clone() username = %v, want %v", cloned.GitUsername, original.GitUsername)
	}

	if cloned.GitEmail != original.GitEmail {
		t.Errorf("Clone() email = %v, want %v", cloned.GitEmail, original.GitEmail)
	}

	if cloned.IsActive != false {
		t.Errorf("Clone() active = %v, want %v", cloned.IsActive, false)
	}

	if cloned.CreatedFrom != "clone" {
		t.Errorf("Clone() createdFrom = %v, want %v", cloned.CreatedFrom, "clone")
	}
}

func TestProfile_detectKeyType(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		want      string
	}{
		{
			name:      "ed25519 key",
			publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILHXUNYnfKa...",
			want:      "id_ed25519",
		},
		{
			name:      "rsa key",
			publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
			want:      "id_rsa",
		},
		{
			name:      "ecdsa key",
			publicKey: "ssh-ecdsa AAAAE2VjZHNhLXNoYTItbmlzdHA...",
			want:      "id_ecdsa",
		},
		{
			name:      "dss key",
			publicKey: "ssh-dss AAAAB3NzaC1kc3MAAACBAI...",
			want:      "id_dsa",
		},
		{
			name:      "unknown key",
			publicKey: "unknown-format key",
			want:      "id_rsa",
		},
		{
			name:      "empty key",
			publicKey: "",
			want:      "id_rsa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{SSHPublicKey: tt.publicKey}
			if got := p.detectKeyType(); got != tt.want {
				t.Errorf("Profile.detectKeyType() = %v, want %v", got, tt.want)
			}
		})
	}
}
