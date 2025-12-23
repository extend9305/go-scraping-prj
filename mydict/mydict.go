package mydict

import "errors"

var errNotFound = errors.New("not found")
var errWordExists = errors.New("word already exists")
var errCantUpdate = errors.New("can't update")
var errCantDelete = errors.New("can't delete")

// Dictionary type
type Dictionary map[string]string

// Search for a word
func (d Dictionary) Search(word string) (string, error) {
	value, exists := d[word]

	if exists {
		return value, nil
	}
	return "", errNotFound
}

// Add  a word to the dictionary
func (d Dictionary) Add(word, def string) error {
	_, err := d.Search(word)
	if err == errNotFound {
		d[word] = def
	} else if err == nil {
		return errWordExists
	}
	return nil
}

// Update a word
func (d Dictionary) Update(word, definition string) error {
	_, err := d.Search(word)
	switch err {
	case nil:
		d[word] = definition
	case errNotFound:
		return errCantUpdate
	}
	return nil
}

// Delete a word
func (d Dictionary) Delete(word string) error {
	_, err := d.Search(word)
	if err == errNotFound {
		return errCantDelete
	}
	delete(d, word)
	return nil
}
