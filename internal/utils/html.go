package utils

import (
	"errors"
	"strings"
)

var ErrHTMLElementNotFound = errors.New("target element not found")

// HTMLCutFirstNode extracts and removes the first occurrence of an HTML element with a specific tag.
// It returns the modified input string, the extracted element, the starting index of the element, and an error (if any).
func HTMLCutFirstNode(input, tag string) (string, string, int, error) {
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	// find opening tag start
	startIndex := strings.Index(input, openTag)
	if startIndex == -1 {
		return input, "", -1, ErrHTMLElementNotFound
	}

	// find opening tag end
	endOpenTag := strings.Index(input[startIndex:], ">")
	if endOpenTag == -1 {
		return input, "", -1, ErrHTMLElementNotFound
	}
	endOpenTag += startIndex

	// find matching close tag
	depth := 1
	searchStart := endOpenTag + 1
	for depth > 0 {
		nextOpenTag := strings.Index(input[searchStart:], openTag)
		nextCloseTag := strings.Index(input[searchStart:], closeTag)

		if nextCloseTag == -1 {
			return input, "", -1, errors.New("malformed HTML: missing closing tag")
		}

		if nextOpenTag != -1 && nextOpenTag < nextCloseTag {
			depth++
			searchStart += nextOpenTag + len(openTag)
		} else {
			depth--
			searchStart += nextCloseTag + len(closeTag)
		}
	}

	endIndex := searchStart
	targetElement := input[startIndex:endIndex]     // extract target
	result := input[:startIndex] + input[endIndex:] // remove target from input
	return result, targetElement, startIndex, nil
}
