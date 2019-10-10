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

package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/xrash/smetrics"
)

const shortestSubStrLen = 3

type lcs struct {
	lengths    [][]int
	aMap, bMap []int
	length     int
	aIndex     int
	bIndex     int
}

func NewLCS(a, b string) lcs {
	l := lcs{
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

func (l *lcs) Next() {
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

func CommonTrigrams(a, b string) float64 {
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
	return float64(c) / float64(len(a)-2+len(b)-2-c)
}

func StringCompare(a, b string) float64 {
	if a == b {
		return 1.0
	}
	return 0.0
}

func Levenshein(a, b string) float64 {
	return 1.0 - float64(smetrics.WagnerFischer(a, b, 1, 1, 2))/
		(float64(len(a)+len(b)))
}

func JaroWinkler(a, b string) float64 {
	return smetrics.JaroWinkler(a, b, 0.7, 4)
}

func LCS(a, b string) float64 {
	s := subStrLen(a, b)
	return float64(s) / float64(len(a)+len(b)-s)
}

func Wrap(f func(a, b string) float64) func(a, b string) float64 {
	return func(a, b string) float64 {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
		r := f(a, b)
		if r < 1.0 && a == b {
			return -1.0
		}
		return r
	}

}

var Sims = map[string]func(a, b string) float64{
	"string compare":  Wrap(StringCompare),
	"levenshein":      Wrap(Levenshein),
	"jaro-winkler":    Wrap(JaroWinkler),
	"lcs":             Wrap(LCS),
	"common trigrams": Wrap(CommonTrigrams),
}

func ListSimilarity(as, bs []string, f func(a, b string) float64) float64 {
	max := 0.0
	for _, a := range as {
		for _, b := range bs {
			if m := f(a, b); m > max {
				max = m
			}
		}
	}
	return max
}

func main() {
	max := 0.0
	maxSim := ""
	for sim, f := range Sims {
		start := time.Now()
		s := 0.0
		for _, n := range Groups {
			i := f(n[0], n[1])
			s += i * i
		}
		for _, a := range Artists {
			i := ListSimilarity(a[0], a[1], f)
			s += i * i

		}
		s = math.Sqrt(s) / math.Sqrt(float64(len(Groups)+len(Artists)))
		fmt.Printf("%5.3f %s took %s\n",
			s, sim, time.Since(start))
		if s > max {
			max = s
			maxSim = sim
		}
	}
	for _, n := range Groups {
		maxScore := Sims[maxSim](n[0], n[1])
		for sim, f := range Sims {
			if sim == maxSim {
				continue
			}
			s := f(n[0], n[1])
			if s > maxScore {
				fmt.Printf(
					"%s(%s,%s) > %s, %5.3f > %5.3f\n",
					sim, n[0], n[1], maxSim, s, maxScore)
			}
		}
	}
	for _, a := range Artists {
		maxScore := ListSimilarity(a[0], a[1], Sims[maxSim])
		for sim, f := range Sims {
			if sim == maxSim {
				continue
			}
			s := ListSimilarity(a[0], a[1], f)
			if s > maxScore {
				fmt.Printf(
					"%s(%s,%s) > %s, %5.3f > %5.3f\n",
					sim,
					strings.Join(a[0], ","),
					strings.Join(a[1], ","),
					maxSim, s, maxScore)
			}
		}
	}
}

var Groups = [][]string{
	{`2019-02-23 Barceló Maya Beach, Riviera Maya, Quintana Roo, Mexico`, `2019-02-23 - Barceló Maya Beach Resort, Riviera Maya, Mexico`},
	{`A Different Kind of Human (Step II)`, `A Different Kind of Human`},
	{`Adrenalin Baby - Johnny Marr Live`, `MARR`},
	{`Anjunadeep 10`, `Anjunadeep 10 Sampler: Part 2`},
	{`Apotheosis, Vol. 1: Mozart - The Final Quartets`, `Mozart: The Final Quartets (apotheosis vol. 1)`},
	{`Back and Forth`, `Back & Forth`},
	{`Back to Mine: Nightmares on Wax`, `Back to Mine`},
	{`Beck-Ola`, `Truth/Beck-Ola`},
	{`Bob Stanley & Pete Wiggs Present Three Day Week (When The Lights Went Out 1972 - 1975)`, `Three Day Week: When The Lights Went Out 1972–1975`},
	{`Brazilliance Vol. 1`, `Brazilliance Vol. 2`},
	{`Brazilliance, Volume 2`, `Brazilliance Volume 1`},
	{`Club Edition Summer 2015`, `Edition 2`},
	{`Complete String Quartets, Vol. 1`, `String Quartets`},
	{`Corail (Remixed)`, `Corail`},
	{`DJ-Kicks`, `DJ-Kicks: Laurel Halo`},
	{`Destination Goa - The Eleventh Chapter`, `Destination`},
	{`Dur Dur of Somalia Volume 1 / Volume 2`, `Dur Dur of Somalia - Volume 1, Volume 2`},
	{`ERR REC Library Vol.2 Science & Technology`, `ERR REC Library Vol​.​2 Science & Technology`},
	{`Episode 1`, `Episode 2`},
	{`Eurobeat Festival Vol. 1`, `Eurobeat Festival Vol. 8`},
	{`Even for just the briefest moment  /  Keep charging this "expiation"  /  Plug in to      making it slightly better`, `Even For Just The Briefest Moment / Keep Charging This “expiation” / Plug In To Making It Slightly Better`},
	{`Everything Not Saved Will Be Lost Part 1 (Remixes)`, `Everything Not Saved Will Be Lost Part 1`},
	{`Everything Not Saved Will Be Lost Part 1`, `Everything Not Saved Will Be Lost Part 1 (Remixes)`},
	{`Everything She Wants`, `Everything She Wants (Remix)`},
	{`FRKWYS Vol. 15: Serenitatem`, `FRKWYS 15: Serenitatem`},
	{`Fantast Remixes, Pt. 2`, `Fantast Remixes, Pt. 3`},
	{`Formations Magnétiques et Phénomènes D’incertitude`, `Formations Magnetiques Et Phenomenes D'incertitude`},
	{`Future Hndrxx Presents The WIZRD`, `Future Hndrxx Presents: The WIZRD`},
	{`Galactic Killer Drums Phaze 4`, `Galactic Killer Drums`},
	{`Heinz Music Best Of Vol. 1`, `Heinz Music Best Of, Vol. 1`},
	{`Jack Le Freak (Extended Remix '87)`, `Le Freak`},
	{`Jesus Christ Superstar - A Rock Opera`, `Jesus Christ Superstar`},
	{`John Wick: Chapter 2`, `John Wick: Chapter 2 (Original Motion Picture Soundtrack)`},
	{`Late Night Tales: Floating Points`, `LateNightTales: Floating Points`},
	{`Let Freedom Ring`, `''Let Freedom Ring''`},
	{`Life of Leaf LP`, `Life of Leaf`},
	{`Live At Carnegie Hall 1977`, `Live At Carnegie Hall`},
	{`Mahler's 1st Symphony In D Major "Titan"`, `Symphony No. 1 "The Titan"`},
	{`Mendelssohn: Violin Concerto in D Minor & String Symphonies Nos. 1-6`, `Violin Concerto in D Minor & String Symphonies Nos. 1-6`},
	{`More Moonglow - The Rock Hard EP`, `Moonglow`},
	{`My Laptops 2001 - 2006`, `My Laptops 2001-2006`},
	{`Mystic Warrior`, `Mystic Warrior EP`},
	{`N9NA Collection 2`, `N9NA`},
	{`Nigeria 70 - No Wahala: Highlife, Afro-Funk & Juju 1973-1987`, `Nigeria 70: No Wahala: Highlife, Afro-Funk & Juju 1973-1987`},
	{`Nova Tunes 3.5`, `Nova Tunes 3.9`},
	{`ONDA (온다)`, `ONDA`},
	{`Pink & Blue (RAC Mix)`, `Pink & Blue`},
	{`Prophecy + Progress`, `Prophecy + Progress: UK Electronics 1978-1990`},
	{`Quentin Tarantino's Once Upon a Time in Hollywood Original Motion Picture Soundtrack`, `Once Upon a Time in Hollywood`},
	{`Remember The Night - Live at EPIC Prague, December 2018`, `Remember the Night (Live at Epic Prague, December 2018)`},
	{`Remind Me (The Classic Elektra Recordings 1978-1984)`, `Remind Me: The Classic Elektra Recordings 1978-1984`},
	{`Road Chronicles Live!`, `Road Chronicles: Live!`},
	{`Samiyam - reflectionz`, `reflectionz`},
	{`Silent Piano: Songs for Sleeping 2`, `Silent Piano (Songs for Sleeping) 2`},
	{`Soul Jazz Records presents CONGO REVOLUTION – Revolutionary and Evolutionary Sounds from the Two Congos 1955-62`, `Congo Revolution: Revolutionary and Evolutionary Sounds from the Two Congos (1955-62)`},
	{`Sounds From The Village Vol. 1`, `Sounds from the Village, Vol. 2`},
	{`Souvenir`, `Souvenir Α CΑRΕΕR ΑΝΤΗΟLΟGΥ 1979-2019`},
	{`Strange Pleasure + New Dawn`, `Strange Pleasure`},
	{`Super Onze - Gao`, `Enregistrés Pour Yehia Le Marabout`},
	{`Symphony in G minor / O Garatuja: Prelude / Série brasileira`, `Symphony in G Minor, O Garatuja Prelude & Série brasileira`},
	{`Symphony no. 1 in D major`, `Symphony no. 4 in E-flat major "Romantic"`},
	{`Tchaikovsky: Symphony No. 4 & Mussorgsky: Pictures at an Exhibition`, `Tchaikovsky: Symphony No. 4 / Mussorgsky: Pictures at an Exhibition`},
	{`The Budos Band`, `The Budos Band V`},
	{`The Legend Lives On`, `The Legend`},
	{`The Valentines Massacre - Amen Project, Pt. 2`, `The Valentines Massacre - Amen Project Pt. 2`},
	{`The Vursiflenze Mismantler`, `The  Vursiflenze Mismantler`},
	{`Two Roomed Hotel`, `Two Roomed Motel`},
	{`Voice Of Resistance (صوت المقاومة)`, `Voice Of Resistance`},
	{`Warehouse 10, Volume 8`, `Warehouse 10 Volume 8`},
	{`marasy collection ～marasy original songs best & new～`, `marasy collection ～marasy original songs best & new`},
}

var Artists = [][][]string{
	{{"Adam Ant"}, {"Adam and The Ants"}},
	{{"Adam and The Ants"}, {"Adam Ant"}},
	{{"Ahmedou Ahmed Lewla"}, {"Ahmedou Ahmed Lowla"}},
	{{"Ahmedou Ahmed Lowla"}, {"Ahmedou Ahmed Lewla"}},
	{{"Airis String Quartet"}, {"Nordic String Quartet"}},
	{{"Beck"}, {"Becky Lamb"}},
	{{"Becky Lamb"}, {"Beck"}},
	{{"Benedictine Monks of Santo Domingo de Silos"}, {"The Benedictine Monks Of Santo Domingo De Silos"}},
	{{"Beyoncé"}, {"Kendrick Lamar", "JAY-Z", "Major Lazer", "WizKid", "Pharrell Williams", "Childish Gambino", "SAINt JHN", "Tiwa Savage", "Burna Boy", "Tekno", "Jessie Reyez", "070 Shake", "Shatta Wale", "Mr. Eazi", "Yemi Alade", "Tierra Whack", "Moonchild Sanelly", "Salatiel"}},
	{{"BiSH (ビッシュ)"}, {"BiSH"}},
	{{"BiSH"}, {"BiSH (ビッシュ)"}},
	{{"Bill Evans (saxophone)"}, {"Bill Evans"}},
	{{"Bill Evans"}, {"Bill Evans (saxophone)"}},
	{{"Biosphere"}, {"biosphere (CA)"}},
	{{"Bonzo Dog Doo/Dah Band"}, {"The Bonzo Dog Band"}},
	{{"Bryan Müller"}, {"Skee Mask"}},
	{{"Bugge Wesseltoft", "Rim Banna", "Checkpoint 303"}, {"Rim Banna (ريم بنا\u200e)"}},
	{{"Call Super"}, {"Ondo Fudd", "The Gathering", "Solex", "Harry Nilsson", "Racoon", "Capricorn", "Speed 78", "Postmen", "Undeclinable Ambuscade", "Onderhonden", "Bloem De Ligny", "Project 2000", "Nuff Said", "Blimey!", "Grof Geschut", "Headfirst", "Gluemen", "Birdskin", "Gitbox!"}},
	{{"Camellia (かめりあ)"}, {"Camellia", "Erasure"}},
	{{"Camellia", "Erasure"}, {"Camellia (かめりあ)"}},
	{{"Ceephax Acid Crew", "SoundLift", "DoubleV", "Afternova", "Nery", "Icone", "Gary Afterlife", "Vax", "ALTIMA", "Costa", "Jamie R", "Dmitry Golban", "Oren", "Moein", "Sequence 11", "Kerris", "Med vs. Neil Bamford"}, {"Ceephax"}},
	{{"Ceephax"}, {"Ceephax Acid Crew", "SoundLift", "DoubleV", "Afternova", "Nery", "Icone", "Gary Afterlife", "Vax", "ALTIMA", "Costa", "Jamie R", "Dmitry Golban", "Oren", "Moein", "Sequence 11", "Kerris", "Med vs. Neil Bamford"}},
	{{"Christian Fennesz"}, {"Fennesz"}},
	{{"Daniel Kandi", "Phillip Alpha"}, {"William Ryan Fritch"}},
	{{"David Essex", "Cockney Rebel", "Hawkwind", "The Kinks", "The Troggs", "Edgar Broughton Band", "Mungo Jerry", "Lieutenant Pigeon", "Matchbox", "Adam Faith", "Phil Cordell", "Bombadil", "Barracuda", "Roly", "The Brothers", "Climax Chicago", "Marty Wilde", "Small Wonder", "Stavely Makepeace", "Ricky Wilde", "Robin Goodfellow", "Mike McGear", "Wigan's Ovation", "Paul Brett", "Stud Leather", "The Sutherland Brothers Band", "The Troll Brothers", "Pheon Bear"}, {"Pete Wiggs", "Bob Stanley", "R. Stevie Moore"}},
	{{"Default (缺省)"}, {"Default", "McCoy Tyner"}},
	{{"Default", "McCoy Tyner"}, {"Default (缺省)"}},
	{{"Dirty Androids"}, {"Machine Head", "Astral Projection", "Human Blue", "NOMA", "Talamasca", "Miranda", "Atmos", "Bamboo Forest", "Tranan", "Aeternum", "Android", "Chromosome", "Saiko-Pod", "Phasio", "Amtrax", "A.I.R."}},
	{{"Double Trouble", "Stevie Ray Vaughan", "Stevie Ray Vaughan & Double Trouble"}, {"Stevie Ray Vaughan and Double Trouble"}},
	{{"Fennesz"}, {"Christian Fennesz"}},
	{{"Frank Iero and the Future Violents"}, {"Frank Iero"}},
	{{"Frank Iero"}, {"Frank Iero and the Future Violents"}},
	{{"Halogenix", "Alix Perez", "Fixate", "Bredren", "Deft", "Tsuruda", "Lewis James", "Monty", "Razat", "Submarine", "Cesco"}, {"Mendo", "Monika Kruse", "Optimuss", "Hollen", "Tom Wax", "Stefano Noferini", "George Privatti", "DJ Dep", "Miguel Bastida", "Yvan Genkins", "Dennis Cruz", "Elio Riso", "Guille Placencia", "Pablo Say", "Neverdogs", "Gianni Firmaio", "Outway", "Anti-Slam & W.E.A.P.O.N.", "M.F.S: Observatory", "Raul Facio", "Danniel Selfmade", "DJ Micky Da Funk", "Costantino Nappi", "Alex Smott", "Natch!", "Dothen", "Simon T", "Frank Storm", "Gesus lpz", "Mr Jefferson", "Paul Darey", "Hannes Bruniic", "Francesco Dinoia", "Sleepy & Boo", "Oxy Beat", "Sonate", "Kostha", "Elio Kenza", "Maris", "Jose Oli", "John Beltran", "Herva", "Delta Funktionen", "Bnjmn", "Bleak"}},
	{{"Haruomi Hosono (細野晴臣)"}, {"Haruomi Hosono"}},
	{{"Haruomi Hosono"}, {"Haruomi Hosono (細野晴臣)"}},
	{{"Heize (헤이즈)"}, {"Heize"}},
	{{"Heize"}, {"Heize (헤이즈)"}},
	{{"Il Gardellino"}, {"Soli & il gardellino"}},
	{{"JAB"}, {"John Also Bennett"}},
	{{"Jambinai (잠비나이)"}, {"Jambinai"}},
	{{"Jambinai"}, {"Jambinai (잠비나이)"}},
	{{"Jan Garbarek - Bobo Stenson Quartet"}, {"Jan Garbarek", "Bobo Stenson"}},
	{{"Jan Garbarek", "Bobo Stenson"}, {"Jan Garbarek - Bobo Stenson Quartet"}},
	{{"Jeezy"}, {"Young Jeezy"}},
	{{"Jeff Beck"}, {"The Jeff Beck Group"}},
	{{"Jerry Garcia Band"}, {"Jerry Garcia", "Jerry Garcia Acoustic Band"}},
	{{"Jerry Garcia", "Jerry Garcia Acoustic Band"}, {"Jerry Garcia Band"}},
	{{"Jim Peterik & World Stage"}, {"Jim Peterik"}},
	{{"Jim Peterik"}, {"Jim Peterik & World Stage"}},
	{{"John Also Bennett"}, {"JAB"}},
	{{"John Diva & the Rockets of Love"}, {"John Diva And The Rockets Of Love"}},
	{{"John Diva And The Rockets Of Love"}, {"John Diva & the Rockets of Love"}},
	{{"John Medeski's Mad Skillet"}, {"John Medeski"}},
	{{"John Medeski"}, {"John Medeski's Mad Skillet"}},
	{{"Johnny Marr"}, {"MARR"}},
	{{"K Á R Y Y N"}, {"KÁRYYN"}},
	{{"K.K. Null"}, {"KK Null"}},
	{{"KK Null"}, {"K.K. Null"}},
	{{"Kalli", "neanderthalic", "Numb Limbs", "Setrus Phree", "Avsluta", "Tweedle", "Thomas 9000", "Haise Pount", "Rites Of Unison"}, {"Funk Fox", "Foamek", "Jacopo SB", "DJ Sagol", "Dj Whipr Snipr", "Nahamasy"}},
	{{"Kalli", "neanderthalic", "Numb Limbs", "Setrus Phree", "Avsluta", "Tweedle", "Thomas 9000", "Haise Pount", "Rites Of Unison"}, {"Pris", "Myler", "Chicago Flotation Device", "Unklon"}},
	{{"Kedr Livanskiy"}, {"Кедр ливанский [Kedr Livanskiy]"}},
	{{"Keiko Osaki"}, {"Keiko"}},
	{{"Keiko"}, {"Keiko Osaki"}},
	{{"Kendrick Lamar", "JAY-Z", "Major Lazer", "WizKid", "Pharrell Williams", "Childish Gambino", "SAINt JHN", "Tiwa Savage", "Burna Boy", "Tekno", "Jessie Reyez", "070 Shake", "Shatta Wale", "Mr. Eazi", "Yemi Alade", "Tierra Whack", "Moonchild Sanelly", "Salatiel"}, {"Beyoncé"}},
	{{"King Gizzard & The Lizard Wizard"}, {"King Gizzard And The Lizard Wizard", "Boris", "Battle of Mice", "Isis", "Melvins", "Sunn O)))", "Mouse on Mars", "The Evens", "Benoît Pioulard", "Dosh", "William Elliott Whitmore", "Bracken", "Boduf Songs", "Michael Cashmore", "French Toast", "Jenny Hoyston", "Joe Lally", "Trencher"}},
	{{"King Gizzard & The Lizard Wizard"}, {"King Gizzard And The Lizard Wizard"}},
	{{"King Gizzard And The Lizard Wizard", "Boris", "Battle of Mice", "Isis", "Melvins", "Sunn O)))", "Mouse on Mars", "The Evens", "Benoît Pioulard", "Dosh", "William Elliott Whitmore", "Bracken", "Boduf Songs", "Michael Cashmore", "French Toast", "Jenny Hoyston", "Joe Lally", "Trencher"}, {"King Gizzard & The Lizard Wizard"}},
	{{"King Gizzard And The Lizard Wizard"}, {"King Gizzard & The Lizard Wizard"}},
	{{"King Sunny Ade & His African Beats"}, {"King Sunny Adé & His African Beats"}},
	{{"King Sunny Adé & His African Beats"}, {"King Sunny Ade & His African Beats"}},
	{{"Korn"}, {"KoЯn"}},
	{{"KoЯn"}, {"Korn"}},
	{{"KÁRYYN"}, {"K Á R Y Y N"}},
	{{"Le Trio Joubran (الثلاثي جبران)"}, {"Trio Joubran"}},
	{{"Lee Perry"}, {"Lee “Scratch” Perry"}},
	{{"Lee “Scratch” Perry"}, {"Lee Perry"}},
	{{"MARR"}, {"Johnny Marr"}},
	{{"Machine Head", "Astral Projection", "Human Blue", "NOMA", "Talamasca", "Miranda", "Atmos", "Bamboo Forest", "Tranan", "Aeternum", "Android", "Chromosome", "Saiko-Pod", "Phasio", "Amtrax", "A.I.R."}, {"Dirty Androids"}},
	{{"Mantra (ES)"}, {"Mantra"}},
	{{"Mantra"}, {"Mantra (ES)"}},
	{{"Marco Passarani"}, {"Passarani"}},
	{{"Mark Ashley"}, {"Few Miles On"}},
	{{"Master Musicians of Jajouka"}, {"The Master Musicians Of Jajouka"}},
	{{"Mekons"}, {"The Mekons"}},
	{{"Mendo", "Monika Kruse", "Optimuss", "Hollen", "Tom Wax", "Stefano Noferini", "George Privatti", "DJ Dep", "Miguel Bastida", "Yvan Genkins", "Dennis Cruz", "Elio Riso", "Guille Placencia", "Pablo Say", "Neverdogs", "Gianni Firmaio", "Outway", "Anti-Slam & W.E.A.P.O.N.", "M.F.S: Observatory", "Raul Facio", "Danniel Selfmade", "DJ Micky Da Funk", "Costantino Nappi", "Alex Smott", "Natch!", "Dothen", "Simon T", "Frank Storm", "Gesus lpz", "Mr Jefferson", "Paul Darey", "Hannes Bruniic", "Francesco Dinoia", "Sleepy & Boo", "Oxy Beat", "Sonate", "Kostha", "Elio Kenza", "Maris", "Jose Oli", "John Beltran", "Herva", "Delta Funktionen", "Bnjmn", "Bleak"}, {"Halogenix", "Alix Perez", "Fixate", "Bredren", "Deft", "Tsuruda", "Lewis James", "Monty", "Razat", "Submarine", "Cesco"}},
	{{"Monari Wakita (脇田もなり)"}, {"Monari Wakita"}},
	{{"Monari Wakita"}, {"Monari Wakita (脇田もなり)"}},
	{{"Nick Cave & The Bad Seeds"}, {"Nick Cave and The Bad Seeds"}},
	{{"Nick Cave and The Bad Seeds"}, {"Nick Cave & The Bad Seeds"}},
	{{"Nordic String Quartet"}, {"Airis String Quartet"}},
	{{"Ola Onabule"}, {"Ola Onabulé"}},
	{{"Ola Onabulé"}, {"Ola Onabule"}},
	{{"Olivier Giacomotto", "Citizen Kain", "Transcode", "Made in Paris", "Malandra Jr.", "Heerhorst"}, {"Agressive Mood", "Sectio Aurea", "Sanathana", "Ozore", "Spagettibrain", "Walpurgisnacht Projekt", "MK-Ultra", "Angkor", "Paradelika", "Naraku", "Sefirot", "Aluxo'Ob", "Sha Manik", "ZY"}},
	{{"Ondo Fudd", "The Gathering", "Solex", "Harry Nilsson", "Racoon", "Capricorn", "Speed 78", "Postmen", "Undeclinable Ambuscade", "Onderhonden", "Bloem De Ligny", "Project 2000", "Nuff Said", "Blimey!", "Grof Geschut", "Headfirst", "Gluemen", "Birdskin", "Gitbox!"}, {"Call Super"}},
	{{"Passarani"}, {"Marco Passarani"}},
	{{"Pat Metheny Group", "Leo Kottke"}, {"Pat Metheny"}},
	{{"Pat Metheny Group"}, {"Pat Metheny"}},
	{{"Pat Metheny"}, {"Pat Metheny Group", "Leo Kottke"}},
	{{"Pat Metheny"}, {"Pat Metheny Group"}},
	{{"Paul McCartney & Wings"}, {"Paul McCartney", "Wings"}},
	{{"Paul McCartney", "Wings"}, {"Paul McCartney & Wings"}},
	{{"Pete Wiggs", "Bob Stanley", "R. Stevie Moore"}, {"David Essex", "Cockney Rebel", "Hawkwind", "The Kinks", "The Troggs", "Edgar Broughton Band", "Mungo Jerry", "Lieutenant Pigeon", "Matchbox", "Adam Faith", "Phil Cordell", "Bombadil", "Barracuda", "Roly", "The Brothers", "Climax Chicago", "Marty Wilde", "Small Wonder", "Stavely Makepeace", "Ricky Wilde", "Robin Goodfellow", "Mike McGear", "Wigan's Ovation", "Paul Brett", "Stud Leather", "The Sutherland Brothers Band", "The Troll Brothers", "Pheon Bear"}},
	{{"Red Velvet (레드벨벳)"}, {"Red Velvet"}},
	{{"Red Velvet"}, {"Red Velvet (레드벨벳)"}},
	{{"Rei Kondoh (近藤嶺)", "Hiroki Morishita (森下弘生)", "Takeru Kanazaki (金崎猛)"}, {"Takeru Kanazaki"}},
	{{"Rim Banna (ريم بنا\u200e)"}, {"Bugge Wesseltoft", "Rim Banna", "Checkpoint 303"}},
	{{"Roland Hanna"}, {"Sir Roland Hanna"}},
	{{"Roni Alter (רוני אלטר)"}, {"Roni Alter"}},
	{{"Roni Alter"}, {"Roni Alter (רוני אלטר)"}},
	{{"Rossington Collins Band"}, {"The Rossington Collins Band"}},
	{{"S.P.Y."}, {"S.P.Y", "Splash"}},
	{{"S.P.Y", "Splash"}, {"S.P.Y."}},
	{{"Sakanaction (サカナクション)"}, {"Sakanaction"}},
	{{"Sakanaction"}, {"Sakanaction (サカナクション)"}},
	{{"Sana (さな)"}, {"sana"}},
	{{"Sir Roland Hanna"}, {"Roland Hanna"}},
	{{"Skee Mask"}, {"Bryan Müller"}},
	{{"Snowy White & The White Flames"}, {"Snowy White And The White Flames"}},
	{{"Snowy White And The White Flames"}, {"Snowy White & The White Flames"}},
	{{"Soli & il gardellino"}, {"Il Gardellino"}},
	{{"Stephane Huchard Cultisong Trio"}, {"Stéphane Huchard"}},
	{{"Stevie Ray Vaughan and Double Trouble"}, {"Double Trouble", "Stevie Ray Vaughan", "Stevie Ray Vaughan & Double Trouble"}},
	{{"Stéphane Huchard"}, {"Stephane Huchard Cultisong Trio"}},
	{{"Super Onze De Gao"}, {"Super Onze"}},
	{{"Super Onze"}, {"Super Onze De Gao"}},
	{{"Suso Sáiz"}, {"Christian Fennesz", "Suso Saiz"}},
	{{"Takeru Kanazaki"}, {"Rei Kondoh (近藤嶺)", "Hiroki Morishita (森下弘生)", "Takeru Kanazaki (金崎猛)"}},
	{{"Tarja Turunen"}, {"Tarja"}},
	{{"Tarja"}, {"Tarja Turunen"}},
	{{"The Benedictine Monks Of Santo Domingo De Silos"}, {"Benedictine Monks of Santo Domingo de Silos"}},
	{{"The Bonzo Dog Band"}, {"Bonzo Dog Doo/Dah Band"}},
	{{"The Cult"}, {"William Ryan Fritch"}},
	{{"The Jeff Beck Group"}, {"Jeff Beck"}},
	{{"The Master Musicians Of Jajouka"}, {"Master Musicians of Jajouka"}},
	{{"The Mekons"}, {"Mekons"}},
	{{"The Rossington Collins Band"}, {"Rossington Collins Band"}},
	{{"Tomeka Reid Quartet"}, {"Tomeka Reid"}},
	{{"Tomeka Reid"}, {"Tomeka Reid Quartet"}},
	{{"Toshifumi Hinata (日向敏文)"}, {"Toshifumi Hinata"}},
	{{"Toshifumi Hinata"}, {"Toshifumi Hinata (日向敏文)"}},
	{{"Trio Joubran"}, {"Le Trio Joubran (الثلاثي جبران)"}},
	{{"Umm Kulthum (أم كلثوم\u200e)"}, {"Umm Kulthum"}},
	{{"Umm Kulthum"}, {"Umm Kulthum (أم كلثوم\u200e)"}},
	{{"Varg (SE)"}, {"Varg"}},
	{{"Varg"}, {"Varg (SE)"}},
	{{"Yorushika (ヨルシカ)"}, {"Yorushika"}},
	{{"Yorushika"}, {"Yorushika (ヨルシカ)"}},
	{{"Young Jeezy"}, {"Jeezy"}},
	{{"Yu Kobayashi", "Yumi Kawamura"}, {"Yumi Kawamura (川村ゆみ)", "Yu Kobayashi (小林ゆう)"}},
	{{"Yumi Kawamura (川村ゆみ)", "Yu Kobayashi (小林ゆう)"}, {"Yu Kobayashi", "Yumi Kawamura"}},
	{{"biosphere (CA)"}, {"Biosphere"}},
	{{"marasy (まらしぃ)"}, {"marasy"}},
	{{"marasy"}, {"marasy (まらしぃ)"}},
	{{"sana"}, {"Sana (さな)"}},
	{{"Кедр ливанский [Kedr Livanskiy]"}, {"Kedr Livanskiy"}},
}
