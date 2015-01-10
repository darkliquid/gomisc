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
func SplitOnVowelGroups(name string) (ret []string) {
	vg := regexp.MustCompile("[AEIOUYaeiouy]+")
	indexes := vg.FindAllStringIndex(name, -1)
	start := 0
	for _, index := range indexes {
		if index[0] > 0 {
			ret = append(ret, name[start:index[1]])
			start = index[1] + 1
		}
	}
	if start < len(name)-1 {
		ret = append(ret, name[start:])
	}
	return
}

func main() {
	gen := 0
	flag.IntVar(&gen, "gen", 0, "generate given number of names")
	write := false
	flag.BoolVar(&write, "w", false, "write out analysis to json files")
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
		vgs := SplitOnVowelGroups(name)
		if len(vgs) > 0 {
			FillGrams(vgs, vowelgroups)
			prefix := vgs[0]
			vgs = vgs[1:]

			if _, ok := prefixes[prefix]; !ok {
				prefixes[prefix] = 1
			} else {
				prefixes[prefix] = prefixes[prefix] + 1
			}

			if len(vgs) > 0 {
				suffix := vgs[len(vgs)-1]
				vgs = vgs[:len(vgs)-1]

				if _, ok := suffixes[suffix]; !ok {
					suffixes[suffix] = 1
				} else {
					suffixes[suffix] = suffixes[suffix] + 1
				}

				if len(vgs) > 0 {
					for _, join := range vgs {
						if _, ok := joins[join]; !ok {
							joins[join] = 1
						} else {
							joins[join] = joins[join] + 1
						}
					}
				}
			}
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

	for i := 0; i < gen; i++ {
		fmt.Println(GenerateVowelGroupName(vowelgroups))
	}
}

// GenerateVowelGroupName makes a name by traversing the vowelgroup randomly.
// It limits the traversal to a maximum of 3 steps and returns immediately on a
// dead end.
func GenerateVowelGroupName(vowelgroups map[string]count) (ret string) {
	key := ""
	if val, ok := vowelgroups[ret]; ok {
		key = val.RandomKey()
	}
	ret = ret + key
	if val, ok := vowelgroups[key]; ok {
		key = val.RandomKey()
	} else {
		return
	}
	ret = ret + key
	if val, ok := vowelgroups[key]; ok {
		key = val.RandomKey()
	} else {
		return
	}
	ret = ret + key
	return
}
