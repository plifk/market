package passwords

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	type args struct {
		password  string
		blacklist []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {
		// 	name:    "password is too short",
		// 	args:    args{password: "short"},
		// 	wantErr: true,
		// },

		// {
		// 	name:    "password is too long",
		// 	args:    args{password: "This password is really long but so easy that you can read it today just once and remember tomorrow morning if you are not tired."},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "repetitive password",
		// 	args:    args{password: strings.Repeat("repeat.", 10)},
		// 	wantErr: true,
		// },
		// {
		// 	name: "password is in blacklist",
		// 	args: args{
		// 		password:  "light-sun-window-led-bulb-galaxy",
		// 		blacklist: []string{"light-sun-window-led-bulb-galaxy"},
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "password is in blacklist (partial, should fail)",
		// 	args: args{password: "light-sun-window-ledx",
		// 		blacklist: []string{"light-sun-window"},
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "password has parts in blacklist but should pass as it's complicated enough",
		// 	args: args{password: "light-sun-window-led-complicated-long",
		// 		blacklist: []string{"light-sun-window"},
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name:    "password is pass-secret",
		// 	args:    args{password: "pass-secret"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password is password123abc",
		// 	args:    args{password: "password123abc"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password contains annoying repetition",
		// 	args:    args{password: "dog-lot-barks-lot-cloud-lot-help-lot"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password contains non-printable chars",
		// 	args:    args{password: "this is invalid: \u0000"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password contains invalid UTF8 sequence",
		// 	args:    args{password: "this is invalid: \xed\xa0\x80\x80"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password contains invalid UTF8 sequence",
		// 	args:    args{password: "this is invalid: \xed\xa0\x80\x80"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password is great",
		// 	args:    args{password: "great-password-is-hard-enough"},
		// 	wantErr: false,
		// },
		// {
		// 	name:    "password is pseudo-hard",
		// 	args:    args{password: "$#!@U*JI($JOI"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password is hard",
		// 	args:    args{password: "$g(U*xI$hJvI"},
		// 	wantErr: false,
		// },
		// {
		// 	name:    "password contains repetitive subset",
		// 	args:    args{password: "ajakaeaabagakajaearabgneakra"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "admin password",
		// 	args:    args{password: "adminadmin"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "root password",
		// 	args:    args{password: "rootadmin!"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password is wordy",
		// 	args:    args{password: "home-joystick-girl-floor-food-nun-nan-loop"},
		// 	wantErr: false,
		// },
		// {
		// 	name:    "password has low entropy",
		// 	args:    args{password: "!!!a+k-z!!!"},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "password has low entropy",
		// 	args:    args{password: "!@#$%ˆ&*()"},
		// 	wantErr: true,
		// },
		{
			name:    "password has enough entropy",
			args:    args{password: "mlernierbngle"},
			wantErr: true,
		},
		// {
		// 	name:    "password has enough entropy",
		// 	args:    args{password: "mnrtiubnn9hnsghi4b"},
		// 	wantErr: false,
		// },
		// {
		// 	name:    "password has enough entropy",
		// 	args:    args{password: "ckRj3b4nCB0m2e"},
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.password, tt.args.blacklist...); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLeaked(t *testing.T) {
	// Most of these passwords are bad and shouldn't be used.
	// Probably this test is never going to be needed, unless the algorithm of Validate is being tweaked for some reason.
	t.Skip("skipping expensive check of leaked passwords")
	filename := "testdata/10-million-passwords.txt"
	if _, err := os.Stat(filename); err != nil {
		cmd := exec.Command("curl", "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/xato-net-10-million-passwords.txt", "-o", filename)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	} else {
		t.Log("using cached passwords file.")
	}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 0
	passed := 0
	co := commonOccurrency{}
	for scanner.Scan() {
		line++
		if password := scanner.Text(); len(password) >= 10 { // Skip the base case.
			if err := Validate(password); err == nil {
				passed++
				fmt.Printf("line %d: %s (valid)\n", line, password)
				co.add(password)
			}
		}
	}
	co.printAll()
	t.Logf("%.2f%% leaked passwords considered valid (%d/%d)", 100*float64(passed)/float64(line), passed, line)
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}

var alphabeticRegex = regexp.MustCompile("[^a-z0-9]+")

// commonOccurrency normalizes and store passwords to help trying to find a trait.
type commonOccurrency map[string]int

func (co commonOccurrency) add(password string) {
	if rest := alphabeticRegex.ReplaceAllString(password, ""); rest != "" {
		co[strings.ToLower(rest)]++
	}
}

func (co commonOccurrency) printAll() {
	var list []string
	for password := range co {
		list = append(list, password)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return co[list[i]] < co[list[j]]
	})
	for _, password := range list {
		fmt.Printf("password containing %q appears %d times\n", password, co[password])
	}
}
