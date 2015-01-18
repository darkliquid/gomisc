package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

type count map[string]uint

// RandomKey returns a random key from the count map, weighted by the counts
func (c *count) RandomKey() string {
	var arr []string
	for key, val := range *c {
		for ; val > 0; val-- {
			arr = append(arr, key)
		}
	}
	if len(arr) == 0 {
		return ""
	}
	if len(arr) == 1 {
		return arr[0]
	}
	return arr[rand.Intn(len(arr)-1)]
}

// NGram builds an ngram list of character groupings from a string
func NGram(name string, size int) (ret []string) {
	tmp := ""
	for index, chr := range []rune(name) {
		tmp = tmp + string(chr)
		if index > 0 && (index+1)%size == 0 {
			ret = append(ret, tmp)
			tmp = ""
		}
	}
	return
}

// FillGrams makes a map of which gram leads to another, weighted by occurence
func FillGrams(gramList []string, gramMap map[string]count) {
	start := ""
	for _, gram := range gramList {
		if _, ok := gramMap[start]; !ok {
			gramMap[start] = make(count)
		}
		if _, ok := gramMap[start][gram]; !ok {
			gramMap[start][gram] = 0
		}
		gramMap[start][gram] = gramMap[start][gram] + 1
		start = gram
	}
}

// SplitOnVowelGroups breaks a string into a chunks on the start of every
// contiguous group of vowels
func SplitOnVowelGroups(name string, bugged bool) (ret []string) {
	vg := regexp.MustCompile("[AEIOUYaeiouy]+")
	indexes := vg.FindAllStringIndex(name, -1)
	start := 0
	for _, index := range indexes {
		if index[0] > 0 {
			ret = append(ret, name[start:index[1]])
			if bugged {
				start = index[1] + 1
			} else {
				start = index[1]
			}
		}
	}
	if len(ret) == 0 {
		ret = append(ret, name)
	}
	if bugged && start <= len(name) && start > 0 {
		if start == len(name) {
			start--
		}
		if len(ret) == 1 {
			ret[0] += name[start:]
		} else {
			ret[len(ret)-1] += name[start:]
		}
	} else if start < len(name) && start > 0 {
		if len(ret) < 2 {
			ret = append(ret, name[start:])
		} else {
			ret[len(ret)-1] += name[start:]
		}
	}
	return
}

func main() {
	min := 0
	flag.IntVar(&min, "m", 0, "minimum size required for a name fragment")
	gen := 0
	flag.IntVar(&gen, "g", 0, "generate given number of names")
	write := false
	flag.BoolVar(&write, "w", false, "write out analysis to json files")
	bugged := false
	flag.BoolVar(&bugged, "b", false, "use bugged implementation of vowel grou splitting")
	stats := false
	flag.BoolVar(&stats, "s", false, "print out analysis stats")
	raw := false
	flag.BoolVar(&raw, "r", false, "print out name parts individually")
	algorithm := "vg3"
	flag.StringVar(&algorithm, "a", "vg3", "generation algorithm [vg3, vg3b, 2gr, 3gr, pt2, pt3]")
	uniq := false
	flag.BoolVar(&uniq, "u", false, "only ever use a name fragment once")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	file, err := os.Open(flag.Arg(0)) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	twograms := make(map[string]count)
	threegrams := make(map[string]count)
	prefixes := make(count)
	joins := make(count)
	suffixes := make(count)
	vowelgroups := make(map[string]count)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if len(name) == 0 {
			continue
		}
		FillGrams(NGram(name, 2), twograms)
		FillGrams(NGram(name, 3), threegrams)
		vgs := SplitOnVowelGroups(name, bugged)
		// Skip processing names that only have a prefix
		if len(vgs) < 2 {
			continue
		}

		FillGrams(vgs, vowelgroups)

		suffix := vgs[len(vgs)-1]
		vgs = vgs[:len(vgs)-1]

		if _, ok := suffixes[suffix]; !ok {
			suffixes[suffix] = 1
		} else {
			suffixes[suffix] = suffixes[suffix] + 1
		}

		for len(vgs) > 1 {
			join := vgs[len(vgs)-1]
			vgs = vgs[:len(vgs)-1]
			if _, ok := joins[join]; !ok {
				joins[join] = 1
			} else {
				joins[join] = joins[join] + 1
			}
		}

		prefix := vgs[len(vgs)-1]
		if _, ok := prefixes[prefix]; !ok {
			prefixes[prefix] = 1
		} else {
			prefixes[prefix] = prefixes[prefix] + 1
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if write {
		// output 2-grams
		b, err := json.Marshal(twograms)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("twograms.json", b, 0755)

		// output 3-grams
		b, err = json.Marshal(threegrams)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("threegrams.json", b, 0755)

		// output prefixes
		b, err = json.Marshal(prefixes)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("prefixes.json", b, 0755)

		// output joins
		b, err = json.Marshal(joins)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("joins.json", b, 0755)

		// output suffixes
		b, err = json.Marshal(suffixes)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("suffixes.json", b, 0755)

		// output vowel groups
		b, err = json.Marshal(vowelgroups)
		if err != nil {
			fmt.Println("error:", err)
		}
		ioutil.WriteFile("vowelgroups.json", b, 0755)
	}

	if stats {
		fmt.Println("Analysis stats:")
		fmt.Printf("  twograms: %d\n", len(twograms))
		fmt.Printf("  threegrams: %d\n", len(threegrams))
		fmt.Printf("  prefixes: %d\n", len(prefixes))
		fmt.Printf("  joins: %d\n", len(joins))
		fmt.Printf("  suffixes: %d\n", len(suffixes))
		fmt.Printf("  vowelgroups: %d\n\n", len(vowelgroups))
	}

	var algFunc func() []string
	switch algorithm {
	case "vg3", "vg3b":
		algFunc = func() []string {
			return GenerateMarkovName(min, vowelgroups, 3, uniq)
		}
	case "2gr":
		algFunc = func() []string {
			return GenerateMarkovName(0, twograms, 6, uniq)
		}
	case "3gr":
		algFunc = func() []string {
			return GenerateMarkovName(0, threegrams, 4, uniq)
		}
	case "pt2":
		algFunc = func() []string {
			return GeneratePartsName(min, uniq, prefixes, suffixes)
		}
	case "pt3":
		algFunc = func() []string {
			return GeneratePartsName(min, uniq, prefixes, joins, suffixes)
		}
	default:
		log.Fatal("Unknown name algorithm specified. Valid algorithms are: vg3, vg3b, 2gr, 3gr, pt2, pt3")
	}

	for i := 0; i < gen; i++ {
		if raw {
			fmt.Println(strings.Join(algFunc(), " "))
		} else {
			fmt.Println(strings.Join(algFunc(), ""))
		}
	}
}

// GenerateMarkovName makes a name by traversing the map randomly.
// It limits the name length to the given maxlen and returns immediately on a
// dead end.
func GenerateMarkovName(min int, markov map[string]count, maxiter int, uniq bool) (ret []string) {
	used := make(map[string]bool, 0)
	key := ""
	if val, ok := markov[""]; ok {
		for i := 0; i < len(markov); i++ {
			key = val.RandomKey()
			if len(key) >= min {
				break
			}
		}
	}
	ret = append(ret, key)
	for i := 0; i < maxiter; i++ {
		if val, ok := markov[key]; ok {
			for i := 0; i < len(markov); i++ {
				key = val.RandomKey()
				if _, ok := used[key]; ok {
					continue
				}
				if len(key) >= min {
					used[key] = true
					break
				}
			}
			ret = append(ret, key)
		} else {
			return
		}
	}
	return
}

// GeneratePartsName makes a name by picking a random item from each list and
// appending it
func GeneratePartsName(min int, uniq bool, lists ...count) (ret []string) {
	used := make(map[string]bool, 0)
	for idx, list := range lists {
		if len(list) == 0 {
			continue
		}
		var keys []string
		for key := range list {
			keys = append(keys, key)
		}
		for i := 0; i < len(keys); i++ {
			j := 0
			if len(keys) > 1 {
				j = rand.Intn(len(keys) - 1)
			}
			key := keys[j]
			if _, ok := used[key]; ok {
				continue
			}
			if len(key) >= min {
				used[key] = true
				ret = append(ret, key)
				if idx == len(lists)-1 {
					return
				}
				break
			}
		}
	}
	return
}
