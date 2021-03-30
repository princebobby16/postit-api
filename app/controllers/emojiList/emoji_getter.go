package emojiList

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"
)

type EmojiPackage struct {
	Emoticons          []string `json:"emoticons"`
	DingBats           []string `json:"ding_bats"`
	Transport          []string `json:"transport"`
	UnCategorized      []string `json:"un_categorized"`
	EnclosedCharacters []string `json:"enclosed_characters"`
}

func HandleGetEmoji(w http.ResponseWriter, _ *http.Request) {

	var emojiPackage EmojiPackage

	emoticons := []int{128513, 128591}
	var sEmoticons []string
	dingbats := []int{9986, 10160}
	var sDingbats []string
	transport := []int{128640, 128704}
	var sTransport []string
	unCategorized := []int{127744, 128512}
	var sUnCategorized []string
	enclosedCharacters := []int{127516, 127570}
	var sEnclosedCharacters []string

	for i := emoticons[0]; i < emoticons[1]; i++ {
		str := html.UnescapeString("&#" + strconv.Itoa(i) + ";")
		sEmoticons = append(sEmoticons, str)
	}

	for i := dingbats[0]; i < dingbats[1]; i++ {
		str := html.UnescapeString("&#" + strconv.Itoa(i) + ";")
		sDingbats = append(sDingbats, str)
	}

	for i := transport[0]; i < transport[1]; i++ {
		str := html.UnescapeString("&#" + strconv.Itoa(i) + ";")
		sTransport = append(sTransport, str)
	}

	for i := unCategorized[0]; i < unCategorized[1]; i++ {
		str := html.UnescapeString("&#" + strconv.Itoa(i) + ";")
		sUnCategorized = append(sUnCategorized, str)
	}

	for i := enclosedCharacters[0]; i < enclosedCharacters[1]; i++ {
		str := html.UnescapeString("&#" + strconv.Itoa(i) + ";")
		sEnclosedCharacters = append(sEnclosedCharacters, str)
	}

	emojiPackage = EmojiPackage{
		Emoticons:          sEmoticons,
		DingBats:           sDingbats,
		Transport:          sTransport,
		UnCategorized:      sUnCategorized,
		EnclosedCharacters: sEnclosedCharacters,
	}

	_ = json.NewEncoder(w).Encode(emojiPackage)
}
