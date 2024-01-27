package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

var logger *log.Logger
var discord *discordgo.Session
var GUILD_ID string
var err error
var API_URL string
var db *sql.DB

type Course struct {
	Crn         string `json:"crn"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Section     string `json:"section"`
	DateRange   any    `json:"dateRange"`
	CourseType  any    `json:"type"`
	Instructor  any    `json:"instructor"`
	SubjectFull string `json:"subjectFull"`
	Subject     string `json:"subject"`
	Campus      string `json:"campus"`
	Comment     any    `json:"comment"`
	Credits     int    `json:"credits"`
	Semester    int    `json:"semester"`
	Level       string `json:"level"`
}

type Time struct {
	Crn       string `json:"crn"`
	Days      string `json:"days"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Location  string `json:"location"`
}

type Seating struct {
	Crn       string `json:"crn"`
	Available string `json:"available"`
	Max       string `json:"max"`
	Waitlist  any    `json:"waitlist"`
	Checked   string `json:"checked"`
}

func errorEmbed(error string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0xff0000,
		Title:  "An error has occurred.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Error",
				Value: error,
			},
		},
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "seatings",
			Description: "Get the seating of a specified course using the Course Registration Number.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "crn",
					Description: "Course Registration Number",
					Required:    true,
				},
			},
		},
		{
			Name:        "courseinfo",
			Description: "Get the info of a course using the Course Registration Number.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "crn",
					Description: "Course Registration Number",
					Required:    true,
				},
			},
		},
		{
			Name:        "searchcourses",
			Description: "Search for CRNs using the course ID.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Course ID (i.e. COMP 1001)",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"seatings": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var embed *discordgo.MessageEmbed

			//calling api here because it already automatically scrapes and does all the dirty work
			request, err := http.Get(API_URL + "/seating?crn=" + i.ApplicationCommandData().Options[0].StringValue())
			if err != nil {
				logger.Fatal(err)
			}
			defer request.Body.Close()
			if request.StatusCode == http.StatusOK {
				//im just gonna hope the code doesnt error
				request1, _ := http.Get(API_URL + "/courses?crn=" + i.ApplicationCommandData().Options[0].StringValue())
				body1, _ := io.ReadAll(request1.Body)
				var course Course
				json.Unmarshal(body1[1:len(body1)-1], &course)
				body, _ := io.ReadAll(request.Body)
				var seating Seating
				json.Unmarshal(body[1:len(body)-1], &seating)
				embed = &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{},
					Color:  0x7f1734,
					Title:  fmt.Sprintf("Seats for %s-%s", course.Id, course.Section),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Total",
							Value: seating.Max,
						},
						{
							Name:  "Available",
							Value: seating.Available,
						},
						{
							Name:  "Waitlist",
							Value: fmt.Sprintf("%v", seating.Waitlist),
						},
					},
					Footer:    &discordgo.MessageEmbedFooter{Text: "ClaretForMUN.com"},
					Timestamp: seating.Checked + ":00-03:30",
				}
			} else {
				embed = errorEmbed("Please double check your CRN")
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			})
		},
		"courseinfo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var embed *discordgo.MessageEmbed

			var course Course
			err := db.QueryRow("SELECT * FROM courses WHERE courses.crn = $1", i.ApplicationCommandData().Options[0].StringValue()).Scan(&course.Crn, &course.Id, &course.Name, &course.Section, &course.DateRange, &course.CourseType, &course.Instructor, &course.Subject, &course.SubjectFull, &course.Campus, &course.Comment, &course.Credits, &course.Semester, &course.Level)
			if err != nil {
				embed = errorEmbed(err.Error())
			} else {
				var comment string
				if course.Comment != nil {
					comment = fmt.Sprintf("%v", course.Comment)
				} else {
					comment = "None"
				}

				var timeStr string

				times, _ := db.Query("SELECT times.days, times.\"startTime\", times.\"endTime\", times.location FROM times WHERE times.crn = $1", course.Crn)
				for times.Next() {
					var time Time
					times.Scan(&time.Days, &time.StartTime, &time.EndTime, &time.Location)
					timeStr += fmt.Sprintf("%s - %s - %s - %s\n", time.Days, time.StartTime, time.EndTime, time.Location)
				}

				embed = &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{},
					Color:  0x7f1734,
					Title:  fmt.Sprintf("%s-%s - %s (%s)", course.Id, course.Section, course.Name, course.Crn),
					URL:    "https://claretformun.com?crns=" + course.Crn,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Campus",
							Value: course.Campus,
						},
						{
							Name:  "Instructor(s)",
							Value: fmt.Sprintf("%v", course.Instructor),
						},
						{
							Name:  "Type",
							Value: fmt.Sprintf("%v", course.CourseType),
						},
						{
							Name:  "Comment", //probably look into hiding this if it the comment doesnt exist
							Value: comment,
						},
						{
							Name:  "Times", //same with times
							Value: timeStr,
						},
					},
					Footer:    &discordgo.MessageEmbedFooter{Text: "ClaretForMUN.com"},
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			})
		},
		"searchcourses": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			var embed *discordgo.MessageEmbed
			var fields []*discordgo.MessageEmbedField

			courses, err := db.Query("SELECT courses.id, courses.section, courses.name, courses.crn, courses.instructor FROM courses WHERE LOWER(courses.id) LIKE LOWER( $1 )", i.ApplicationCommandData().Options[0].StringValue()+"%")
			if err != nil {
				embed = errorEmbed(err.Error())
			} else {
				for courses.Next() {
					logger.Println("still exists?")

					var course Course
					courses.Scan(&course.Id, &course.Section, &course.Name, &course.Crn, &course.Instructor)
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:  fmt.Sprintf("%s-%s - %s (%s)", course.Id, course.Section, course.Name, course.Crn),
						Value: fmt.Sprintf("Instructor(s): %v", course.Instructor),
					})

					embed = &discordgo.MessageEmbed{
						Author:    &discordgo.MessageEmbedAuthor{},
						Color:     0x7f1734,
						Title:     fmt.Sprintf("Search for %s", i.ApplicationCommandData().Options[0].StringValue()),
						Fields:    fields,
						Footer:    &discordgo.MessageEmbedFooter{Text: "ClaretForMUN.com"},
						Timestamp: time.Now().Format(time.RFC3339),
					}
				}
			}

			if len(fields) > 25 {
				embed = errorEmbed("Too many courses were found, if possible please specify your search term.")
			}

			logger.Println("still exists?")

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			})
		},
	}
)

func init() {
	logger = log.Default()
	logger.Println("👋 Claret Discord Bot")

	TOKEN := os.Getenv("TOKEN")
	if TOKEN == "" {
		logger.Fatal("TOKEN is not defined in environment variables")
	}

	GUILD_ID = os.Getenv("GUILD_ID")
	if GUILD_ID == "" {
		logger.Fatal("GUILD_ID is not defined in environment variables")
	}

	DB_URL := os.Getenv("DB_URL")
	if GUILD_ID == "" {
		logger.Fatal("DB_URL is not defined in environment variables")
	}

	API_URL = os.Getenv("API_URL")
	if GUILD_ID == "" {
		logger.Fatal("API_URL is not defined in environment variables")
	}

	discord, err = discordgo.New("Bot " + TOKEN)
	if err != nil {
		logger.Fatal(err)
	}

	db, err = sql.Open("pgx", DB_URL)
	if err != nil {
		logger.Fatal(err)
	}
}

func main() {
	defer db.Close()

	err := db.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("💿 Connected to Database!")

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Printf("✅ Logged in as %v#%v!", s.State.User.Username, s.State.User.Discriminator)
	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = discord.Open()
	if err != nil {
		logger.Fatal(err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, GUILD_ID, v)
		if err != nil {
			logger.Fatalf("❌ Cannot create command '%v': %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	logger.Println("✅ Added commands!")

	defer discord.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	logger.Println("🔌 Shutting down bot...")

	//mainly for testing
	if slices.Contains(os.Args, "--removeCommands") {
		logger.Println("🗑️  Removing commands...")
		for _, v := range registeredCommands {
			err := discord.ApplicationCommandDelete(discord.State.User.ID, v.GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}
