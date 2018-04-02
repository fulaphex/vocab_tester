Simple script in `go` for revising stuff (e.g. vocabulary).  
Progress is saved in `json` file.

## How to use
`go run vocab.go` to run the script  
in each line you'll see query followed by the number of possible answers  
`la cena (3):`  
after you press enter you'll see all possible answers and you're prompter to 
type `y` if you answered correctly or `n` if you didn't know the answer and 
confirm with enter

## Changing the vocabulary
Supplied `vocab.csv` corresponds to some basic spanish vocabulary. It contains 
two values per row, `:` used as separator. 
