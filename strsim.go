// Copyright © 2018 Charles Haynes <ceh@ceh.bz>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package strsim

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/xrash/smetrics"
)

const shortestSubStrLen = 3

type LCS struct {
	lengths    [][]int
	aMap, bMap []int
	length     int
	aIndex     int
	bIndex     int
}

func NewLCS(a, b string) LCS {
	l := LCS{
		lengths: make([][]int, len(a)),
		aMap:    make([]int, len(a)),
		bMap:    make([]int, len(b)),
	}
	for i := range l.aMap {
		l.aMap[i] = i
	}
	for i := range l.bMap {
		l.bMap[i] = i
	}
	for i := range l.lengths {
		l.lengths[i] = make([]int, len(b))
		for j := range l.lengths[i] {
			if a[i] != b[j] {
				continue
			}
			if i == 0 || j == 0 {
				l.lengths[i][j] = 1
			} else {
				l.lengths[i][j] = l.lengths[i-1][j-1] + 1
			}
			if l.lengths[i][j] > l.length {
				l.length = l.lengths[i][j]
				l.aIndex = i + 1
				l.bIndex = j + 1
			}
		}
	}
	return l
}

func (l *LCS) Next() {
	ml := 0
	ai := 0
	bi := 0
	l.lengths = append(l.lengths[:l.aIndex-l.length],
		l.lengths[l.aIndex:]...)
	amai := l.aMap[l.aIndex-1] + 1
	bmbi := l.bMap[l.bIndex-1] + 1
	l.aMap = append(l.aMap[:l.aIndex-l.length], l.aMap[l.aIndex:]...)
	l.bMap = append(l.bMap[:l.bIndex-l.length], l.bMap[l.bIndex:]...)
	for i := range l.lengths {
		l.lengths[i] = append(l.lengths[i][:l.bIndex-l.length],
			l.lengths[i][l.bIndex:]...)
		for j := range l.lengths[i] {
			if l.aMap[i] >= amai {
				if l.aMap[i]-l.lengths[i][j] < amai {
					l.lengths[i][j] = l.aMap[i] - amai + 1
				}
			}
			if l.bMap[j] >= bmbi {
				if l.bMap[j]-l.lengths[i][j] < bmbi {
					l.lengths[i][j] = l.bMap[j] - bmbi + 1
				}
			}
			if l.lengths[i][j] > ml {
				ml = l.lengths[i][j]
				ai = i + 1
				bi = j + 1
			}
		}
	}
	l.length = ml
	l.aIndex = ai
	l.bIndex = bi
}

func subStrLen(a, b string) int {
	if len(a) < shortestSubStrLen || len(b) < shortestSubStrLen {
		if a == b {
			return len(a)
		}
		return 0
	}
	r := 0
	for l := NewLCS(a, b); l.length >= shortestSubStrLen; l.Next() {
		r += l.length
	}
	return r
}

func common3grams(a, b string) float64 {
	if len(a) < 3 || len(b) < 3 {
		if a == b {
			return 1.0
		}
		return 0.0
	}
	tg := map[string]int{}
	for i := 3; i <= len(a); i++ {
		tg[a[i-3:i]]++
	}
	c := 0
	for i := 3; i <= len(b); i++ {
		if tg[b[i-3:i]] > 0 {
			c++
			tg[b[i-3:i]]--
		}
	}
	return float64(c) /
		math.Min(float64(len(a)-2), float64(len(b)-2))
}

var similarity = map[string]func(a, b string) float64{
	"string compare": func(a, b string) float64 {
		if a == b {
			return 1.0
		}
		return 0.0
	},
	"levenshein": func(a, b string) float64 {
		return 1.0 - float64(smetrics.WagnerFischer(a, b, 1, 1, 2))/
			(float64(len(a)+len(b)))
	},
	"jaro-winkler": func(a, b string) float64 {
		return smetrics.JaroWinkler(a, b, 0.7, 4)
	},
	"lcs": func(a, b string) float64 {
		return float64(subStrLen(a, b)) /
			math.Min(float64(len(a)), float64(len(b)))
	},
	"common 3grams": func(a, b string) float64 {
		return common3grams(a, b)
	},
}

func Similarity(a, b, sim string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	r := similarity[sim](a, b)
	if r < 1.0 && a == b {
		return -1.0
	}
	return r
}

func main() {
	sims := map[string]float64{
		"string compare": 0.0,
		"levenshein":     0.0,
		"jaro-winkler":   0.0,
		"lcs":            0.0,
		"common 3grams":  0.0,
	}
	for sim := range sims {
		start := time.Now()
		for _, n := range names {
			sims[sim] += Similarity(n.a, n.b, sim)
		}
		for _, a := range artists {
			sims[sim] += ArtistsSimilarity(a.a, a.b, sim)

		}
		fmt.Printf("%s took %s, avg %5.3f\n", sim, time.Since(start),
			sims[sim]/float64(len(names)+len(artists)))
	}
	max := 0.0
	s := ""
	for sim, v := range sims {
		if v > max {
			s = sim
			max = v
		}
	}
	for _, n := range names {
		ns := Similarity(n.a, n.b, s)
		for sim := range sims {
			if sim == s {
				continue
			}
			rs := Similarity(n.a, n.b, sim)
			if rs > ns {
				fmt.Printf(
					"%s(%s,%s) > %s, %5.3f > %5.3f\n",
					sim, n.a, n.b, s, rs, ns)
			}
		}
	}
	for _, a := range artists {
		as := ArtistsSimilarity(a.a, a.b, s)
		for sim := range sims {
			if sim == s {
				continue
			}
			rs := ArtistsSimilarity(a.a, a.b, sim)
			if rs > as {
				fmt.Printf(
					"%s(%s,%s) > %s, %5.3f > %5.3f\n",
					sim,
					strings.Join(a.a, ","),
					strings.Join(a.b, ","),
					s, rs, as)
			}
		}
	}
}
