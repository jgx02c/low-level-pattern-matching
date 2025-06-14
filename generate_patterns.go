//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strconv"
)

// Base patterns for generating variations
var basePatterns = []string{
	"he said",
	"she said",
	"he told",
	"she told",
	"i heard",
	"they said",
	"someone said",
	"according to",
	"reportedly",
	"allegedly",
	"sources say",
	"witnesses claim",
	"testimony indicates",
	"plaintiff claims",
	"defendant stated",
	"witness testified",
	"court records show",
	"evidence suggests",
	"attorney argued",
	"counsel stated",
}

// Prefixes and suffixes to create variations
var prefixes = []string{
	"", "apparently ", "clearly ", "obviously ", "supposedly ", "allegedly ",
	"reportedly ", "presumably ", "evidently ", "seemingly ", "ostensibly ",
	"purportedly ", "conceivably ", "potentially ", "possibly ", "probably ",
	"likely ", "certainly ", "definitely ", "undoubtedly ", "surely ",
}

var suffixes = []string{
	"", " that", " yesterday", " today", " recently", " earlier", " before",
	" during the meeting", " in court", " under oath", " in the deposition",
	" to the jury", " to the judge", " to counsel", " to the witness",
	" in the record", " on the stand", " in testimony", " in evidence",
	" in the filing", " in the brief", " in the motion", " in the pleading",
}

var subjects = []string{
	"the defendant", "the plaintiff", "the witness", "the attorney", "the judge",
	"the jury", "the expert", "the doctor", "the officer", "the investigator",
	"the client", "the victim", "the suspect", "the accused", "the complainant",
	"the respondent", "the petitioner", "the appellant", "the appellee",
	"the party", "the individual", "the person", "the entity", "the corporation",
}

var verbs = []string{
	"was", "were", "had", "did", "would", "could", "should", "might", "may",
	"will", "shall", "must", "can", "cannot", "won't", "wouldn't", "couldn't",
	"shouldn't", "didn't", "hadn't", "hasn't", "haven't", "isn't", "aren't",
}

var objects = []string{
	"guilty", "innocent", "liable", "responsible", "negligent", "fraudulent",
	"compliant", "non-compliant", "present", "absent", "aware", "unaware",
	"informed", "uninformed", "cooperative", "uncooperative", "truthful",
	"dishonest", "credible", "incredible", "reliable", "unreliable",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_patterns.go <number_of_patterns>")
		fmt.Println("Example: go run generate_patterns.go 1000000")
		os.Exit(1)
	}

	numPatterns, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: Invalid number: %v\n", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("patterns_%d.txt", numPatterns)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Printf("🏗️  Generating %d legal hearsay patterns...\n", numPatterns)

	// Write header
	file.WriteString("# Generated Legal Hearsay Detection Patterns\n")
	file.WriteString(fmt.Sprintf("# Total patterns: %d\n", numPatterns))
	file.WriteString("# Generated for performance testing\n\n")

	generated := 0
	patternSet := make(map[string]bool) // To avoid duplicates

	// Generate patterns using combinations
	for generated < numPatterns {
		for _, base := range basePatterns {
			if generated >= numPatterns {
				break
			}

			for _, prefix := range prefixes {
				if generated >= numPatterns {
					break
				}

				for _, suffix := range suffixes {
					if generated >= numPatterns {
						break
					}

					// Create basic pattern
					pattern := prefix + base + suffix
					pattern = cleanPattern(pattern)

					if !patternSet[pattern] && len(pattern) > 3 {
						file.WriteString(pattern + "\n")
						patternSet[pattern] = true
						generated++

						if generated%100000 == 0 {
							fmt.Printf("📝 Generated %d patterns...\n", generated)
						}
					}

					// Create variations with subjects and verbs
					if generated < numPatterns {
						for _, subject := range subjects {
							if generated >= numPatterns {
								break
							}

							for _, verb := range verbs {
								if generated >= numPatterns {
									break
								}

								for _, object := range objects {
									if generated >= numPatterns {
										break
									}

									// Pattern: "prefix + base + that + subject + verb + object"
									complexPattern := fmt.Sprintf("%s%s that %s %s %s", prefix, base, subject, verb, object)
									complexPattern = cleanPattern(complexPattern)

									if !patternSet[complexPattern] && len(complexPattern) > 10 {
										file.WriteString(complexPattern + "\n")
										patternSet[complexPattern] = true
										generated++

										if generated%100000 == 0 {
											fmt.Printf("📝 Generated %d patterns...\n", generated)
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// If we haven't reached the target, add numbered variations
		if generated < numPatterns {
			for i := 0; generated < numPatterns; i++ {
				for _, base := range basePatterns {
					if generated >= numPatterns {
						break
					}

					pattern := fmt.Sprintf("%s %d", base, i)
					if !patternSet[pattern] {
						file.WriteString(pattern + "\n")
						patternSet[pattern] = true
						generated++

						if generated%100000 == 0 {
							fmt.Printf("📝 Generated %d patterns...\n", generated)
						}
					}
				}
			}
		}
	}

	fmt.Printf("✅ Successfully generated %d unique patterns in %s\n", generated, filename)
	fmt.Printf("📊 File size: %.2f MB\n", float64(getFileSize(filename))/(1024*1024))
	fmt.Printf("🚀 Ready for performance testing!\n")
}

func cleanPattern(pattern string) string {
	// Remove extra spaces
	result := ""
	lastWasSpace := false

	for _, char := range pattern {
		if char == ' ' {
			if !lastWasSpace {
				result += " "
				lastWasSpace = true
			}
		} else {
			result += string(char)
			lastWasSpace = false
		}
	}

	// Trim leading/trailing spaces
	if len(result) > 0 && result[0] == ' ' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}

	return result
}

func getFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}
