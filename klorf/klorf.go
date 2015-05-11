package klorf

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	irc "github.com/klorf/goirc/client"
)

type Klorf struct {
	Logger  string `json:"logger"`
	Channel string `json:"channel"`
}

func NewKlorf() *Klorf {
	k := new(Klorf)

	return k
}

func New(log string) *Klorf {
	k := new(Klorf)
	k.Logger = log

	return k
}

func (k *Klorf) Roll(conn *irc.Conn, line *irc.Line) {
	c := line.Args[0]

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m := tokens(line.Args[1])
	d := regexp.MustCompile(`(\d+)d(\d+)([\-\+\*\/]?)(\d*)`)
	msg := ""
	matched := false

	for _, y := range m {
		if d.MatchString(y) {
			matched = true

			roll, err := k.runRoll(d.FindStringSubmatch(y), r)
			if err != nil {
				conn.Privmsg(c, fmt.Sprintf("%s: %s", line.Nick, err.Error()))
			}
			msg = fmt.Sprintf("%s %s", msg, roll)
		} else {
			msg = fmt.Sprintf("%s %s", msg, y)
		}
	}

	if matched {
		msg = fmt.Sprintf("%s:%s", line.Nick, msg)

		chanLog := string(line.Args[0][1:])
		if fmt.Sprintf("%q", chanLog[0]) == "'#'" {
			chanLog = chanLog[1:]
		}

		f := fmt.Sprintf("%s%s_%d-%d-%d.txt", k.Logger, chanLog, line.Time.Year(), line.Time.Month(), line.Time.Day())
		k.logToFile(f, msg)

		conn.Privmsg(c, msg)
	}
}

func (k *Klorf) Log(conn *irc.Conn, line *irc.Line) {
	entry := fmt.Sprintf("%s: %s", line.Nick, line.Args[1])

	c := string(line.Args[0][1:])
	if fmt.Sprintf("%q", c[0]) == "'#'" {
		c = c[1:]
	}

	f := fmt.Sprintf("%s%s_%d-%d-%d.txt", k.Logger, c, line.Time.Year(), line.Time.Month(), line.Time.Day())
	k.logToFile(f, entry)
}

func (k *Klorf) runRoll(in []string, r *rand.Rand) (string, error) {
	var err error
	msg := "["
	total := 0

	diceCount, _ := strconv.Atoi(in[1])
	if diceCount < 1 {
		return "", errors.New("Too little die")
	} else if diceCount > 30 {
		err = errors.New("Setting maximum die to 30")
		diceCount = 30
	}

	diceType, _ := strconv.Atoi(in[2])
	if diceType < 4 || diceType > 20 || diceType%2 != 0 {
		return "", errors.New("Invalid dice type")
	}

	for i := 0; i < diceCount; i++ {
		roll := r.Intn(diceType) + 1

		total += roll
		msg = fmt.Sprintf("%s %d", msg, roll)
	}
	msg = fmt.Sprintf("%s ]", msg)

	t, _ := strconv.Atoi(in[4])
	if in[3] == "+" {
		msg = fmt.Sprintf("%s + %s", msg, in[4])
		total += t
	} else if in[3] == "-" {
		msg = fmt.Sprintf("%s - %s", msg, in[4])
		total -= t
	}
	msg = fmt.Sprintf("(%s = %d)", msg, total)

	return msg, err
}

func (k *Klorf) logToFile(f, m string) {
	fh, _ := os.OpenFile(f, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0775)
	defer fh.Close()

	lfile := log.New(fh, "", log.LstdFlags)
	lfile.Println(m)
}

func tokens(m string) []string {
	return strings.Split(m, " ")
}
