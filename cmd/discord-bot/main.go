package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "(Optional) Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "(Required) Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	Help           = flag.Bool("help", false, "Print help")
)

var s *discordgo.Session

func init() {
	flag.Parse()
}

func errorRespond(s *discordgo.Session, i *discordgo.InteractionCreate, errorString string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// Ignore type for now, they will be discussed in "responses"
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Nick is bad at coding... (" + errorString + ")",
		},
	})
}

var (
	integerOptionMinValue = 2.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "random-sequence",
			Description: "Random sequence of integers from 1 to the number you select",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "The number of random numbers to generate a sequence of (2-100)",
					MinValue:    &integerOptionMinValue,
					MaxValue:    100,
					Required:    true,
				},
			},
		},
		// {
		// 	Name:        "randomized-user-list",
		// 	Description: "Random sequence of integers from 1 to the number you select.",
		// 	Options: []*discordgo.ApplicationCommandOption{
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionChannel,
		// 			Name:        "which-channel",
		// 			Description: "Which Channel do you want to generate this for?",
		// 			// Channel type mask
		// 			ChannelTypes: []discordgo.ChannelType{
		// 				discordgo.ChannelTypeGuildVoice,
		// 				discordgo.ChannelTypeGuildText,
		// 			},
		// 			Required: true,
		// 		},
		// 	},
		// },
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// "slynickel": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 	// Access options in the order provided by the user.
		// 	options := i.ApplicationCommandData().Options
		// 	if len(options) != 1 {
		// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 			// Ignore type for now, they will be discussed in "responses"
		// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 			Data: &discordgo.InteractionResponseData{
		// 				Content: fmt.Sprintf("Internal Error, Nick is bad at coding... (option list length is %d not 1)", len(options)),
		// 			},
		// 		})
		// 		return
		// 	}
		// 	channelChoosen := options[0]
		// 	ch := channelChoosen.ChannelValue(s)

		// 	s2, err := s.Channel(ch.ID)

		// 	if err != nil {
		// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 			// Ignore type for now, they will be discussed in "responses"
		// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 			Data: &discordgo.InteractionResponseData{
		// 				Content: fmt.Sprintf("Internal Error, Nick is bad at coding... getting channel info: %v", err),
		// 			},
		// 		})
		// 		return
		// 	}

		// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 		// Ignore type for now, they will be discussed in "responses"
		// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 		Data: &discordgo.InteractionResponseData{
		// 			Content: fmt.Sprintf("%+v", s2),
		// 		},
		// 	})
		// },
		"random-sequence": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// If the option list isn't one something is wrong
			if len(options) != 1 {
				errorRespond(s, i, fmt.Sprintf("option list length is %d not 1", len(options)))
				return
			}

			// If option type isn't int something is wrong,
			// *ApplicationCommandInteractionDataOption.IntValue() panics if the data type is wrong
			//  we want to avoid that.
			if options[0].Type != discordgo.ApplicationCommandOptionInteger {
				errorRespond(s, i, "IntValue called on data option of type "+options[0].Type.String())
				return
			}
			seqNumber := options[0].IntValue()

			// Generate permutation 0,seqNumber
			m := rand.Perm(int(seqNumber))

			// Convert 0 index array to 1 index array
			strArray := make([]string, len(m))
			for i, j := range m {
				strArray[i] = strconv.Itoa(j + 1)
			}
			log.Printf("Logged sequence: %s", strings.Join(strArray, ","))
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your pseudo-random sequence:\n" + strings.Join(strArray, "\n"),
				},
			})
		},
	}
)

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	if *BotToken == "" {
		flag.PrintDefaults()
		log.Fatal("Bot token not set")
		return
	}

	if *Help {
		flag.PrintDefaults()
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Printf("Created command '%v'", v.Name)
		registeredCommands[i] = cmd
	}

	defer s.Close()

	if *RemoveCommands {
		log.Print("rmcmd flag set, upon exit commands will be unregistered")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
