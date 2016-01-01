package main

import (
	"fmt"
	"log"
	"os"
	"io"
	"io/ioutil"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
	_ "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	database, err := CreateDatabase("clues.db")
	defer database.db.Commit()
	if err != nil {
		log.Fatal(err)
	}
	files, _ := ioutil.ReadDir("./j-archive/")
	for i := 1; i < len(files); i++ {
		fmt.Printf("Parsing game #%d\n", i)
		contents, _ := open_file(fmt.Sprintf("./j-archive/%d.html", i))
		game := parse_page(contents, i)
		fmt.Printf("Writing game #%d\n", i)
		database.InsertGame(game)
	}
}

func open_file(path string) (file io.Reader, err error) {
	file, err = os.Open(path)
	if err != nil {
		return
	}
	return
}

// Parse 
func parse_page(page io.Reader, game_id int) (game Game) {
	game = Game{
		id: game_id,
		airdate: "",
		categories: []Category{},
	}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		panic(err)
	}
	doc.Find(".category").Each(func(i int, s *goquery.Selection) {
		category_name := s.Find(".category_name").Text()
		category_desc := s.Find(".category_comments").Text()
		new_category := Category{
			name: category_name,
			description: category_desc,
			clues: []Clue{},
		}
		game.categories = append(game.categories, new_category)
	})
	clue_iter := 0
	doc.Find(".clue").Each(func(i int, s *goquery.Selection) {
		clue_text := s.Find(".clue_text").Text()
		// Handle empty clues.
		if len(clue_text) == 0 {
			return
		}
		numerator := i
		category := i % 6
		round := i / 30 + 1
		if i >= 30 {
			numerator -= 30
			category += 6
		}
		position := (numerator / 6) + 1
		daily_double := strings.Contains(s.Find(".clue_header").Text(), "DD")
		// Handle final clue.
		if i == 60 {
			s = doc.Find(".final_round")
			round = 3
			category = 12
		}
		mouseover, _ := s.Find("div").Attr("onmouseover")
		r, _ := regexp.Compile(`response\\{0,1}\">(.*?)<\/em>`)
		match := r.FindString(mouseover)
		if len(match) > 0 {
			r, _ = regexp.Compile("<[^>]*>")
			match = r.ReplaceAllString(match, "")[10:]
			match = strings.Replace(match, ">", "", 1)
		}
		/*fmt.Printf("%d [%s] (c %d, r %d, p %d) %s\n",
			i, match, category, round, position, clue_text)*/
		current_clue := Clue{
			clue: clue_text,
			answer: match,
			position: position,
			round: round,
			bonus: daily_double,
		}
		game.categories[category].clues = append(game.categories[category].clues, current_clue)
		clue_iter++
	})
	return
}
