package passwords

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxPasswordLength accepted by the system.
const MaxPasswordLength = 128

// Validate if password seems to follow guidelines, such as
// NIST Special Publication Digital Identity Guidelines https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-63b.pdf
// and OWASP recommendations https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
//
// Other approaches:
// https://haveibeenpwned.com/ provides an API for leaked passwords.
// zxcvbn: Low-Budget Password Strength Estimation
// https://github.com/nbutton23/zxcvbn-go
// Talk: https://www.usenix.org/conference/usenixsecurity16/technical-sessions/presentation/wheeler
func Validate(password string, unsafe ...string) error {
	const minPasswordLength = 10

	// The longer the password, the harder it is for a hashing algorithm to process it. Therefore, we want to limit it too.
	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password length should be at most %d", MaxPasswordLength)
	}
	if len(password) < minPasswordLength {
		return fmt.Errorf("password length should be at least %d", minPasswordLength)
	}
	if !utf8.ValidString(password) {
		return errors.New("password should be valid UTF8 string")
	}

	var (
		entropy     = map[rune]int{}
		last, last2 rune
		sequence    int
		numbers     int
		letters     int
	)
	lp := strings.ToLower(password)
	for _, c := range lp {
		cQwerty, ok := qwertyShiftDigit[c]
		if !ok {
			cQwerty = c
		}
		if !unicode.IsGraphic(c) {
			return errors.New("password should only contain printable chars")
		}
		if unicode.IsDigit(c) {
			numbers++
		}
		if unicode.IsLetter(c) {
			letters++
		}
		entropy[c]++
		// Check just if it is the very same char previously used or the one UTF-8 char immediately before/after
		// or one of the nearest key on a QWERTY keyboard.
		if c == last || c == last-1 || c == last+1 || c == last2 || c == last2-1 || c == last2+1 {
			sequence++
		} else if near, ok := qwertyTable[cQwerty]; ok {
			for _, n := range near {
				lq, lq2 := last, last2
				if q, ok := qwertyShiftDigit[last]; ok {
					lq = q
				}
				if q, ok := qwertyShiftDigit[last2]; ok {
					lq2 = q
				}
				if n == lq || n == lq2 {
					sequence++
				}
			}
		}
		last2 = last
		last = c
	}
	// Try to avoid user with passwords like "bigstring!" as replacing a single letter for a digit or special char is a common trait.
	if letters == len(lp)-1 && letters < minPasswordLength+4 && !unicode.IsLetter(rune(lp[0])) && !unicode.IsLetter(rune(lp[len(lp)-1])) {
		return errLowEntropy
	}
	const minChars = 5 // A password requires at least 5 different runes.
	if numbers > int(0.8*float64(len(lp))) ||
		len(entropy) < minChars ||
		(len(lp) < 14 && sequence >= 3 && letters+numbers > 9) ||
		(len(lp) < 15 && sequence > 7) ||
		(len(lp) < 16 && sequence > 9 && letters > 12) ||
		(3*sequence > 2*len(lp)) {
		return errLowEntropy
	}
	var freq []int
	for _, repeated := range entropy {
		freq = append(freq, repeated)
	}
	sort.Ints(freq)
	// Check if same runes are used above a certain threshold:
	twoTopRunes := float64(freq[0] + freq[1])
	fiveTopRunes := float64(freq[0] + freq[1] + freq[2] + freq[3] + freq[4])
	lenPassword := float64(len(lp))
	if twoTopRunes >= lenPassword*0.5 || fiveTopRunes >= lenPassword*0.6 {
		return errLowEntropy
	}
	// Check if the least used runes are used a lot:
	fiveLeastUsedFreq := freq[len(freq)-1] + freq[len(freq)-2] + freq[len(freq)-3] + freq[len(freq)-4] + freq[len(freq)-5]
	if float64(fiveLeastUsedFreq)-2 > lenPassword*0.7 {
		return errLowEntropy
	}
	for n := 0; n < len(lp)-3; n++ {
		if strings.Count(password, password[n:n+3]) >= 4 {
			return errLowEntropy
		}
	}
	unsafe = append(unsafe, defaultUnsafe...)
	for _, badword := range unsafe {
		badword = strings.ToLower(badword)
		if lp == badword || (len(badword) >= 4 && len(lp)-letters < 3 && strings.Contains(lp, badword)) {
			return errLowEntropy
		}
	}
	return nil
}

var defaultUnsafe = []string{"pass", "password", "p4ss", "p4ssw0rd", "secret", "senha", "love", "iloveyou", "ronaldo", "computer", "money", "12345", "54321"}

var errLowEntropy = errors.New("password has low entropy")

var qwertyTable = map[rune][]rune{}

func init() {
	qwertyTable = neighbours()
}

func neighbours() map[rune][]rune {
	all := map[rune][]rune{}
	for y := range qwerty {
		for x, v := range qwerty[y] {
			all[v] = neighboursPos(x, y)
		}
	}
	return all
}

// Qwerty layout approximated to a rectangle.
var qwerty = [][]rune{
	{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
	{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p'},
	{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';'},
	{'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/'},
}

var qwertyShiftDigit = map[rune]rune{
	'!': '1', '@': '2',
	'#': '3', '$': '4',
	'%': '5', '^': '6',
	'&': '7', '*': '8',
	'(': '9', ')': '0',
}

func neighboursPos(x, y int) (near []rune) {
	horizontal := len(qwerty[0])
	vertical := len(qwerty)
	for w := x - 1; w <= x+1; w++ {
		if w < 0 || w >= horizontal {
			continue
		}
		for h := y - 1; h <= y+1; h++ {
			if h < 0 || h >= vertical {
				continue
			}
			near = append(near, qwerty[h][w])
		}
	}
	return near
}
