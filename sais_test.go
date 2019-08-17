package suffixarray

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
)

func NaiveBuildSuffixArray(text string) []uint64 {
	suffixes := make([]string, len(text)+1)
	for i := 0; i < len(text); i++ {
		suffixes[i] = text[i:]
	}
	sort.Strings(suffixes)
	suffixArray := make([]uint64, len(suffixes))
	for i, suffix := range suffixes {
		suffixArray[i] = uint64(len(text) - len(suffix))
	}
	return suffixArray
}

func TestTypeMap_Build(t *testing.T) {
	type testrow struct {
		Input    string
		Expected string
	}
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		for i, row := range []testrow{
			testrow{"abcdefg", "[SSSSSSL@]"},
			testrow{"habcdefg", "[L@SSSSSL@]"},
			testrow{"rikki-tikki-tikka", "[L@LLL@L@LLL@L@LLL@]"},
			testrow{banana, bananaTM},
			testrow{banana2, banana2TM},
			testrow{cabbage, cabbageTM},
			testrow{loremIpsum, loremIpsumTM},
			testrow{abcdefgh, abcdefghTM},
			testrow{aaaaaaaa, aaaaaaaaTM},
		} {
			text := NewTextFromString(row.Input, opts...)

			typeMap, err := BuildTypeMap(text, opts...)
			if err != nil {
				t.Errorf("[%s/%03d] BuildTypeMap %q: error: %v", cfg.Name, i, row.Input, err)
				continue
			}

			actual := typeMap.Debug()

			if row.Expected != actual {
				t.Errorf("[%s/%03d] BuildTypeMap %q: expected %q, got %q", cfg.Name, i, row.Input, row.Expected, actual)
			}
		}
	}
}

func TestTypeMap_SubstringEq(t *testing.T) {
	type testrow struct {
		I, J     uint64
		Expected bool
	}
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		input := "rikki-tikki-tikka"
		text := NewTextFromString(input, opts...)

		typeMap, err := BuildTypeMap(text, opts...)
		if err != nil {
			t.Errorf("[%s] BuildTypeMap %q: error: %v", cfg.Name, input, err)
			continue
		}
		for i, row := range []testrow{
			testrow{1, 7, true},
			testrow{1, 13, false},
		} {
			actual, err := typeMap.LMSSubstringsAreEqual(text, row.I, row.J)
			if err != nil {
				t.Errorf("[%s/%03d] LMSSubstringsAreEqual %d, %d: expected %v, got %v", cfg.Name, i, row.I, row.J, row.Expected, actual)
				continue
			}
			if row.Expected != actual {
				t.Errorf("[%s/%03d] LMSSubstringsAreEqual %d, %d: expected %v, got %v", cfg.Name, i, row.I, row.J, row.Expected, actual)
			}
		}
	}
}

func TestBuildBucketSizes(t *testing.T) {
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		opts = extendOptions(opts,
			NumValues(7),
			BytesPerValue(1))

		text, err := NewText(7, opts...)
		if err != nil {
			t.Errorf("[%s] NewText: error: %v", cfg.Name, err)
			continue
		}

		textIter := text.Iterate(0, 7)
		for textIter.Next() {
			index := textIter.Index()
			ch := cabbage[index]
			textIter.SetSymbol(uint64(ch - 'a'))
		}
		if err := textIter.Close(); err != nil {
			t.Errorf("[%s] Iterate: error: %v", cfg.Name, err)
			break
		}

		bucketSizes, err := BuildBucketSizes(text)
		if err != nil {
			t.Errorf("[%s] BuildBucketSizes: error: %v", cfg.Name, err)
			continue
		}

		actual := fmt.Sprintf("%v", bucketSizes)
		expected := "[2 2 1 0 1 0 1]"
		if expected != actual {
			t.Errorf("[%s] BuildBucketSizes: expected %v, got %v", cfg.Name, expected, actual)
		}

		bucketHeads := BuildBucketHeads(bucketSizes)
		actual = fmt.Sprintf("%v", bucketHeads)
		expected = "[1 3 5 6 6 7 7]"
		if expected != actual {
			t.Errorf("[%s] BuildBucketHeads: expected %v, got %v", cfg.Name, expected, actual)
		}

		bucketTails := BuildBucketTails(bucketSizes)
		actual = fmt.Sprintf("%v", bucketTails)
		expected = "[2 4 5 5 6 6 7]"
		if expected != actual {
			t.Errorf("[%s] BuildBucketTails: expected %v, got %v", cfg.Name, expected, actual)
		}
	}
}

func TestBuildSuffixArray(t *testing.T) {
	type testrow struct {
		Input    string
		Expected string
	}
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		for i, row := range []testrow{
			testrow{"abcdefg", "[7 0 1 2 3 4 5 6]"},
			testrow{"gfedcba", "[7 6 5 4 3 2 1 0]"},
			testrow{"baabaabac", "[9 1 4 2 5 7 0 3 6 8]"},
			testrow{banana, bananaSA},
			testrow{cabbage, cabbageSA},
			testrow{loremIpsum, loremIpsumSA},
			testrow{abcdefgh, abcdefghSA},
			testrow{aaaaaaaa, aaaaaaaaSA},
		} {
			text := NewTextFromString(row.Input, opts...)

			sa, err := BuildSuffixArray(text, opts...)
			if err != nil {
				t.Errorf("[%s/%03d] BuildSuffixArray %q: error: %v", cfg.Name, i, row.Input, err)
				continue
			}

			canon := NaiveBuildSuffixArray(row.Input)
			computed := fmt.Sprintf("%v", canon)
			if row.Expected != computed {
				t.Errorf("[%s/%d] BuildSuffixArray %q: have the wrong answer saved; we expect %v, but %v is the right answer", cfg.Name, i, row.Input, row.Expected, computed)
			}

			var buf bytes.Buffer
			buf.WriteByte('[')
			sa.ForEach(func(index uint64, pos uint64) error {
				if index > 0 {
					buf.WriteByte(' ')
				}
				fmt.Fprintf(&buf, "%d", pos)
				return nil
			})
			buf.WriteByte(']')
			actual := buf.String()
			if row.Expected != actual {
				t.Errorf("[%s/%03d] BuildSuffixArray %q: expected %v, got %v", cfg.Name, i, row.Input, row.Expected, actual)
			}
		}
	}
}
