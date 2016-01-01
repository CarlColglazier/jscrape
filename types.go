package main

type Clue struct {
	clue string
	answer string
	position int
	round int
	bonus bool
}

type Category struct {
	name string
	description string
	clues []Clue
}

type Game struct {
	id int
	airdate string
	categories []Category
}

