package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"math"
	"regexp"
	"net"
	"net/http"
)

const (
	spam = "spam"
	noSpam = "no spam"
)

type wordFrequency struct {
	word    string
	counter map[string]int
}

type classifier struct {
	dataset map[string][]string
	words   map[string]wordFrequency
}

func readData(path string) map[string]string {
	f, _ := http.Get(path)
	defer f.Body.Close()

	dataset := make(map[string]string)
	textComplete := bufio.NewScanner(f.Body)
	for textComplete.Scan() {
		line := textComplete.Text()
		data := strings.Split(line, "\t")
		
		sentence := data[0]
		if data[1] == "0" {
			dataset[sentence] = noSpam
		} else if data[1] == "1" {
			dataset[sentence] = spam
		}
	}
	return dataset
}

func controller(conex net.Conn){
	defer conex.Close()
	rIn := bufio.NewReader(conex)
	sentence, _ := rIn.ReadString('\n')

	fmt.Println("Llego el mensaje: ", sentence)
	
	nClass := newClassifier()
	dataset := readData("https://raw.githubusercontent.com/alonzoDiego/SpamData/master/spam01.txt")
	//dataset := readData("https://raw.githubusercontent.com/alonzoDiego/SpamData/master/spam02.txt")

	nClass.trainData(dataset)
		
	nClass.classifyInput(sentence)
}

func main() {
	ln, err := net.Listen("tcp","localhost:8000")
	if err != nil {
		fmt.Println("Falla al momento de establecer comunicacion con el puerto 8000", err.Error)
		os.Exit(1)
	}
	defer ln.Close()

	conex, er := ln.Accept()
	if err != nil {
		fmt.Println("Falla al momento de aceptar Cliente", er.Error)
	}
	controller(conex)
}

func newClassifier() *classifier {
	cls := new(classifier)
	cls.dataset = map[string][]string{
		spam: []string{},
		noSpam: []string{},
	}
	cls.words = map[string]wordFrequency{}
	return cls
}

func (cls *classifier) trainData(dataset map[string]string) {
	for sentence, class := range dataset {
		cls.addSentence(sentence, class)
		words := tokenize(sentence)
		for _, w := range words {
			cls.addWord(w, class)
		}
	}
}

func (cls *classifier) addSentence(sentence, class string) {
	cls.dataset[class] = append(cls.dataset[class], sentence)
}

func (cls *classifier) addWord(word, class string) {
	wf, ok := cls.words[word]
	if !ok {
		wf = wordFrequency{word: word, counter: map[string]int{
			spam: 0,
			noSpam: 0,
		}}
	}
	wf.counter[class]++
	cls.words[word] = wf
}

func (cls classifier) classifyInput(sentence string) {
	words := tokenize(sentence)
	spamProb := cls.probability(words, spam)
	noSpamProb := cls.probability(words, noSpam)
	result := make(map[string]float64)

	result[spam] = spamProb
	result[noSpam] = noSpamProb
	
	class := ""
	if result[spam] > result[noSpam] {
		class = spam
	} else {
		class = noSpam
	}
	fmt.Printf("...El mensaje es %s\n\n", class)
}

func (cls classifier) probability(words []string, class string) float64 {
	prob := cls.priorProb(class)
	for _, w := range words {
		count := 0
		if wf, ok := cls.words[w]; ok {
			count = wf.counter[class]
		}
		prob *= (float64((count + 1)) / float64((cls.totalWordCount(class) + cls.totalDistinctWordCount())))
	}
	for _, w := range words {
		count := 0
		if wf, ok := cls.words[w]; ok {
			count += (wf.counter[spam] + wf.counter[noSpam])
		}
		prob /= (float64((count + 1)) / float64((cls.totalWordCount("") + cls.totalDistinctWordCount())))
	}
	return prob
}

func (cls classifier) priorProb(class string) float64 {
	return float64(len(cls.dataset[class])) / float64(len(cls.dataset[spam])+len(cls.dataset[noSpam]))
}

func (cls classifier) totalWordCount(class string) int {
	spamCount := 0
	noSpamCount := 0
	for _, wf := range cls.words {
		spamCount += wf.counter[spam]
		noSpamCount += wf.counter[noSpam]
	}
	if class == spam {
		return spamCount
	} else if class == noSpam {
		return noSpamCount
	} else {
		return spamCount + noSpamCount
	}
}

func (cls classifier) totalDistinctWordCount() int {
	spamCount := 0
	noSpamCount := 0
	for _, wf := range cls.words {
		spamCount += zeroOneTransform(wf.counter[spam])
		noSpamCount += zeroOneTransform(wf.counter[noSpam])
	}
	return spamCount + noSpamCount
}

func tokenize(sentence string) []string {
	s := format(sentence)
	words := strings.Fields(s)
	var tokens []string
	for _, w := range words {
		if !isStopword(w) {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

func format(sentence string) string {
	re := regexp.MustCompile("[^a-zA-Z 0-9]+")
	return re.ReplaceAllString(strings.ToLower(sentence), "")
}

func zeroOneTransform(x int) int {
	return int(math.Ceil(float64(x) / (float64(x) + 1.0)))
}

func isStopword(w string) bool {
	_, ok := stopwords[w]
	return ok
}

var stopwords = map[string]struct{}{
	"yo": struct{}{}, "mi": struct{}{}, "propio": struct{}{}, "nosotros": struct{}{}, "nuestro": struct{}{},
	"tu": struct{}{}, "tus": struct{}{}, "tuyo": struct{}{},
	"él": struct{}{}, "su": struct{}{}, "el": struct{}{}, "ella": struct{}{},
	"esto": struct{}{}, "ellos": struct{}{}, "ustedes": struct{}{},
	"qué": struct{}{}, "cuál": struct{}{}, "quién": struct{}{}, "cuáles": struct{}{},
	"que": struct{}{}, "ello": struct{}{}, "aquello": struct{}{}, "soy": struct{}{}, "es": struct{}{}, "son": struct{}{}, "fue": struct{}{},
	"fueron": struct{}{}, "por": struct{}{}, "siendo": struct{}{}, "tiene": struct{}{}, "tienes": struct{}{}, "tuvo": struct{}{},
	"teniendo": struct{}{}, "hacer": struct{}{}, "haces": struct{}{}, "hizo": struct{}{}, "haciendo": struct{}{}, "a": struct{}{}, "un": struct{}{},
	"la": struct{}{}, "y": struct{}{}, "pero": struct{}{}, "si": struct{}{}, "o": struct{}{}, "porque": struct{}{},
	"hasta": struct{}{}, "mientras": struct{}{}, "con": struct{}{},
	"sobre": struct{}{}, "contra": struct{}{}, "entre": struct{}{}, "dentro": struct{}{}, "mediante": struct{}{}, "durante": struct{}{},
	"antes": struct{}{}, "después": struct{}{}, "encima": struct{}{}, "detras": struct{}{}, "arriba": struct{}{},
	"abajo": struct{}{}, "en": struct{}{}, "afuera": struct{}{}, "apagado": struct{}{}, "terminado": struct{}{}, "debajo": struct{}{},
	"lejos": struct{}{}, "desde": struct{}{}, "aqui": struct{}{}, "ahí": struct{}{}, "cuándo": struct{}{},
	"dónde": struct{}{}, "cómo": struct{}{}, "todo": struct{}{}, "nada": struct{}{}, "ambos": struct{}{}, "cada": struct{}{},
	"poco": struct{}{}, "más": struct{}{}, "ninguno": struct{}{}, "otro": struct{}{}, "igual": struct{}{}, "tal": struct{}{},
	"pocos": struct{}{}, "todos": struct{}{}, "solo": struct{}{}, "iguales": struct{}{}, "entonces": struct{}{}, "aveces": struct{}{}, "también": struct{}{},
	"muy": struct{}{}, "puede": struct{}{}, "podrás": struct{}{}, "justo": struct{}{}, "puedo": struct{}{}, "debería": struct{}{}, "debe": struct{}{},
	"ahora": struct{}{}, "luego": struct{}{}, "no": struct{}{}, "tuviste": struct{}{},
	"querer": struct{}{}, "queria't": struct{}{}, "quieres": struct{}{},
}