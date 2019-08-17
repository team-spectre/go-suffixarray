package suffixarray

import (
	"fmt"
	"testing"
)

const sampleText = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque tristique,
risus at hendrerit varius, mauris lacus consequat urna, eu porttitor nunc
sapien sit amet orci. Duis iaculis malesuada molestie. Phasellus nec lorem sit
amet dolor tincidunt semper id at felis. Pellentesque non lectus nisi. In
feugiat in justo in gravida. Etiam non elementum mauris. Sed varius, mi vel
gravida condimentum, enim metus vestibulum orci, sed sodales odio dolor in
nunc. Nulla in scelerisque lacus. Orci varius natoque penatibus et magnis dis
parturient montes, nascetur ridiculus mus.

Cras varius arcu sed felis mattis, quis blandit dui vulputate. Proin dictum
nunc ultrices dui euismod, vitae tincidunt ex molestie. Morbi mattis placerat
nulla sed aliquet. Nulla gravida, justo dignissim semper maximus, dolor metus
bibendum nisi, et finibus dui felis ut quam. Donec porttitor, mauris nec
hendrerit condimentum, odio quam ornare ante, vitae tristique nunc tellus non
lorem. Nunc tincidunt magna at dui feugiat, ut placerat enim lacinia. Phasellus
id lacus luctus purus pulvinar ornare. Etiam vel ante luctus, dictum eros eget,
tristique enim. Phasellus nec sapien risus. Sed placerat vel odio vel accumsan.

Maecenas fringilla viverra arcu, sit amet fermentum quam iaculis non. Ut nec
nisl vel massa posuere auctor. Etiam massa dolor, placerat id nibh vitae,
feugiat tempor mi. Nullam rutrum elit mi, sed rutrum quam hendrerit eget. In
hac habitasse platea dictumst. Vivamus eget lobortis metus. Donec dignissim
tempus suscipit. Aenean est lacus, iaculis id hendrerit sed, semper eget neque.

Aenean neque massa, aliquet eget arcu vel, lobortis faucibus est. Sed imperdiet
lacus at laoreet tempus. Curabitur semper nec mi tempus sagittis. Suspendisse
in risus id risus consequat egestas. Aliquam aliquam suscipit auctor. Curabitur
diam lacus, fringilla sit amet laoreet quis, aliquet id ipsum. Duis ornare at
ipsum nec facilisis. Maecenas pulvinar risus at lacus commodo rhoncus. Donec
arcu felis, dictum sit amet porttitor sed, interdum id libero. Nullam quis
dolor ligula.

Morbi venenatis vehicula velit quis tempus. Ut euismod tellus cursus, venenatis
arcu ut, posuere odio. Mauris vitae diam nunc. Curabitur ipsum justo, egestas
vitae dapibus nec, vulputate ac ipsum. Etiam lacinia neque at quam consectetur
condimentum. Fusce nec risus luctus, maximus mi vel, laoreet ex. Sed tristique
facilisis mauris id bibendum.
`

const searchPhrase = `odio`

func TestSearch_EndToEnd(t *testing.T) {
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		text := NewTextFromString(sampleText, opts...)

		sa, err := BuildSuffixArray(text, opts...)
		if err != nil {
			t.Errorf("[%s] BuildSuffixArray: error: %v", cfg.Name, err)
			return
		}

		lcp, err := BuildLCPArray(text, sa, opts...)
		if err != nil {
			t.Errorf("[%s] BuildLCPArray: error: %v", cfg.Name, err)
			return
		}

		lcplr, err := BuildLCPLRArray(lcp, opts...)
		if err != nil {
			t.Errorf("[%s] BuildLCPLRArray: error: %v", cfg.Name, err)
			return
		}

		offsets, err := Search(text, sa, lcplr, searchPhrase)
		if err != nil {
			t.Errorf("[%s] Search: error: %v", cfg.Name, err)
			return
		}

		expected := "[441 905 1181 2166]"
		actual := fmt.Sprintf("%v", offsets)
		if expected != actual {
			t.Errorf("[%s] Search: expected %s, got %s", cfg.Name, expected, actual)
		}
	}
}
