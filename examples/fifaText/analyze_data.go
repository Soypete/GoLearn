package fifa

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kniren/gota/dataframe"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// UnderstandData pulls raw data from Kaggle world cup twitter data set.
// Gleans important insights from data. Publish graphs and print out statistical insights.
// Make rough data set from text data with lable of text or not.
func UnderstandData(file string) error {
	tweets, err := getData(file)
	if err != nil {
		return fmt.Errorf("error getting tweets %v", err)
	}
	df := dataframe.LoadStructs(tweets)
	fmt.Println(df)
	fmt.Println(df.Select([]string{"Length", "Likes", "Retweets", "Followers"}).Describe())

	plotNums(df)
	plotWords(df)
	return nil
}

func plotNums(df dataframe.DataFrame) error {
	colNames := []string{"Length", "Likes", "Retweets", "Followers"}
	for _, colName := range colNames {
		var plotVals plotter.Values
		for _, floatVal := range df.Col(colName).Float() {
			plotVals = append(plotVals, floatVal)
		}

		p, err := plot.New()
		if err != nil {
			return err
		}
		p.Title.Text = fmt.Sprintf("Histogram of a Tweet %s", colName)

		// Create a histogram of our values drawn
		// from the standard normal.
		h, err := plotter.NewHist(plotVals, 16)
		if err != nil {
			return err
		}
		// Normalize the area under the histogram to
		// sum to one.
		p.Add(h)
		if err := p.Save(4*vg.Inch, 4*vg.Inch, "graphs/"+colName+"_hist.png"); err != nil {
			return err
		}
	}
	return nil
}
func plotWords(df dataframe.DataFrame) error {
	var data []WordData
	mapCounts := make(map[string]int, 100000)

	// make map of names that will be labeled as true

	// there are two columns that contain name data. WE need to take names from
	// both columns
	colNames := []string{"Name", "Usermention"}
	for _, colName := range colNames {
		vals := df.Col(colName).Records()
		for _, val := range vals {
			// split multi word names into single words
			words := strings.Split(val, " ")
			for _, w := range words {
				// remove spaces
				w = strings.Trim(w, " ")
				if _, ok := mapCounts[w]; !ok {
					mapCounts[w] = 1
					continue
				}
				mapCounts[w]++
			}
		}
		// add data to struct that contains important data information.
		isName := true
		for name, count := range mapCounts {
			d := WordData{
				Word:       name,
				Occurances: count,
				IsName:     isName,
			}
			data = append(data, d)
		}
	}
	// parse each word in the tweets and lable it as name or not
	records := df.Col("OrgTweet").Records()
	tweetWordsCount := make(map[string]int, 100000)
	for _, r := range records {
		words := strings.Split(r, " ")
		for _, w := range words {
			w = strings.Trim(w, " ")
			// skip names that we already have
			if mapCounts[w] > 0 {
				continue
			}
			// we assume we have all the names. This is going
			// cause some un realiability in the data.
			if _, ok := tweetWordsCount[w]; !ok {
				tweetWordsCount[w] = 1
				continue
			}
			tweetWordsCount[w]++
		}
	}
	// add stata to struct
	for name, count := range tweetWordsCount {
		d := WordData{
			Word:       name,
			Occurances: count,
			IsName:     false,
		}
		data = append(data, d)
	}
	fmt.Printf("From the Fifa World Cup Data you have parsed out %d unique words. \n", len(data))

	// create a new data frame
	newDF := dataframe.LoadStructs(data)
	fmt.Println(newDF)
	// get summary of word occurances
	fmt.Println(newDF.Select([]string{"Occurances"}).Describe())
	// cache data frame in csv
	myFile, err := os.Create("data/words.csv")
	if err != nil {
		return err
	}
	err = newDF.WriteCSV(myFile)
	if err != nil {
		return err
	}
	return nil
}

func getData(file string) ([]Tweet, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	var tweets []Tweet
	for i, line := range lines {
		if i == 0 {
			continue
		}
		// fmt.Println(line)
		id, _ := strconv.Atoi(line[0])
		layout := "2006-01-02"
		d, _ := time.Parse(layout, line[2])
		l, _ := strconv.Atoi(line[4])
		likes, _ := strconv.Atoi(line[7])
		r, _ := strconv.Atoi(line[8])
		f, _ := strconv.Atoi(line[14])
		friends, _ := strconv.Atoi(line[14])

		t := Tweet{
			ID:            id,
			Lang:          line[1],
			Date:          d.Unix(),
			Source:        line[3],
			Length:        l,
			OrgTweet:      line[5],
			Tweets:        line[6],
			Likes:         likes,
			Retweets:      r,
			Hashtag:       line[9],
			Usermention:   line[10],
			UserMentionID: line[11],
			Name:          line[12],
			Place:         line[13],
			Followers:     f,
			Friends:       friends,
		}
		// fmt.Println(line[6])
		tweets = append(tweets, t)
	}
	return tweets, err
}
