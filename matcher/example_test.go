package matcher_test

import (
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/matcher"
	"log"
	"time"
)

func ExampleBasic_Match() {
	basicMatcher, err := matcher.NewBasic("", "", "asset\\d+\\.txt", nil)
	if err != nil {
		log.Fatal(err)
	}
	matched := basicMatcher.Match("parent location", file.NewInfo("asset001.txt", 20, 0644, time.Now(), false))
	fmt.Printf("matched: %v\n", matched)

}
