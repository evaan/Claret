package main

import (
	"log"
	"os"
	"os/signal"
	"slices"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var logger *log.Logger
var discord *discordgo.Session
var GUILD_ID string
var err error

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "test-command",
			Description: "Test command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "string",
					Description: "String Option",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"test-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "it works yippee!",
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
	if DB_URL == "" {
		logger.Fatal("DB_URL is not defined in environment variables")
	}

	discord, err = discordgo.New("Bot " + TOKEN)
	if err != nil {
		logger.Fatal(err)
	}
}

func main() {
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
			logger.Fatalf("EMOJI HERE FUCK YOU Cannot create command '%v': %v", v.Name, err)
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
