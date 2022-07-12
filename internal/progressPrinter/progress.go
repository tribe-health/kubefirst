package progressPrinter

import (
    "sync"
	"github.com/jedib0t/go-pretty/v6/progress"
	"flag"
	"time"
	"fmt"
)

type ActionTracker struct {
	Tracker *progress.Tracker
}

type progressPrinter struct {
	Trackers map[string]*ActionTracker
	pw progress.Writer
}


var instance *progressPrinter
var once sync.Once


func GetInstance() *progressPrinter {
    once.Do(func() {
        instance = &progressPrinter{}
		instance.Trackers = make(map[string]*ActionTracker)
    })
    return instance
}

//Initialise a tracker object
func CreateTracker(title string, total int64) *progress.Tracker {
	tracker := &progress.Tracker{
		Message: title,
		Total:   total,
		Units:   progress.UnitsDefault,
	}

	instance.pw.AppendTracker(tracker)
	return tracker
}

// SetupProgress prepare the progress bar setting its initial configuration
func SetupProgress(numTrackers int) {
	flag.Parse()
	fmt.Printf("Init actions: %d expected tasks ...\n\n", numTrackers)
	// instantiate a Progress Writer and set up the options
	instance.pw = progress.NewWriter()
	instance.pw.SetAutoStop(false)
	instance.pw.SetTrackerLength(30)
	instance.pw.SetMessageWidth(29)
	instance.pw.SetNumTrackersExpected(numTrackers)
	instance.pw.SetSortBy(progress.SortByPercentDsc)
	instance.pw.SetStyle(progress.StyleDefault)
	instance.pw.SetTrackerPosition(progress.PositionRight)
	instance.pw.SetUpdateFrequency(time.Millisecond * 100)
	instance.pw.Style().Colors = progress.StyleColorsExample
	instance.pw.Style().Options.PercentFormat = "%4.1f%%"
	instance.pw.Style().Visibility.ETA = true
	instance.pw.Style().Visibility.ETAOverall = true
	instance.pw.Style().Visibility.Percentage = true
	instance.pw.Style().Visibility.Time = true
	instance.pw.Style().Visibility.TrackerOverall = true
	instance.pw.Style().Visibility.Value = true
	go instance.pw.Render()
}

func LogMessage(message string){
	instance.pw.Log(message)
}

//Add Tracker 
// Return a string for the key
func AddTracker(key string, title string, total int64) string {
	instance.Trackers[key] = &ActionTracker{Tracker: CreateTracker(title, total)}
	return key
}

func IncrementTracker(key string,  value int64) {
	instance.Trackers[key].Tracker.Increment(int64(1))
}
