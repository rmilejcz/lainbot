package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lainbot/cmd"
	"github.com/lainbot/framework"
)

var (
	conf       *framework.Config
	CmdHandler *framework.CommandHandler
	Sessions   *framework.SessionManager
	youtube    *framework.Youtube
	botId      string
	PREFIX     string
)

var buffer = make([][]byte, 0)

func init() {
	conf = framework.LoadConfig("config.json")
	PREFIX = conf.Prefix
}

func main() {
	err := loadSound()
	if err != nil {
		log.Println("Error loading sound: ", err)
		log.Println("Please copy $GOPATH/src/github.com/bwmarrin/examples/airhorn/airhorn.dca to this directory.")
		return
	}
	Sessions = framework.NewSessionManager()
	CmdHandler = framework.NewCommandHandler()
	youtube = &framework.Youtube{Conf: conf}
	registerCommands()
	dg, err := discordgo.New("Bot " + conf.BotToken)
	if err != nil {
		panic(err.Error())
	}
	//dg.Identify.Intents = discordgo.IntentsAll
	dg.AddHandler(ready)
	dg.AddHandler(commandHandler)
	//dg.AddHandler(messageCreate)
	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}
	defer dg.Close()
	log.Println("airhorn is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func recoverErr() {
	if err := recover(); err != nil {
		log.Println(err)
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println("got message")
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "!airhorn") {

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					log.Println("Error playing sound:", err)
				}

				return
			}
		}
	}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("ready")
	s.UpdateGameStatus(0, "games with ur heart")
}

func commandHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	user := message.Author
	if user.ID == botId || user.Bot {
		return
	}
	content := message.Content
	if len(content) <= len(PREFIX) {
		return
	}
	if content[:len(PREFIX)] != PREFIX {
		return
	}
	content = content[len(PREFIX):]
	if len(content) < 1 {
		return
	}

	args := strings.Fields(content)
	if !strings.HasPrefix(args[0], "!") {
		return
	}
	name := strings.ToLower(strings.Replace(args[0], "!", "", 1))
	print(name)
	command, found := CmdHandler.Get(name)
	if !found {
		return
	}
	channel, err := discord.State.Channel(message.ChannelID)
	if err != nil {
		log.Println("Error getting channel,", err)
		return
	}
	guild, err := discord.State.Guild(channel.GuildID)
	if err != nil {
		log.Println("Error getting guild,", err)
		return
	}
	ctx := framework.NewContext(discord, guild, channel, user, message, conf, CmdHandler, Sessions, youtube)
	ctx.Args = args[1:]
	c := *command
	c(*ctx)
}

func loadSound() error {

	file, err := os.Open("airhorn.dca")
	if err != nil {
		log.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			log.Println("Error reading from dca file :", err)
			return err
		}

		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		if err != nil {
			log.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

func registerCommands() {
	CmdHandler.GetCmds()
	CmdHandler.Register("help", cmd.HelpCommand, "Gives you this help message!")
	print("registering commands\n")
	CmdHandler.Register("join", cmd.JoinCommand, "Join a voice channel !join attic")
	print("registering commands\n")
	CmdHandler.Register("leave", cmd.LeaveCommand, "Leaves current voice channel")
	print("registering commands\n")
	CmdHandler.Register("play", cmd.PlayCommand, "Plays whats in the queue")
	print("registering commands\n")
	CmdHandler.Register("stop", cmd.StopCommand, "Stops the music")
	print("registering commands\n")
	CmdHandler.Register("info", cmd.InfoCommand, "???")
	print("registering commands\n")
	CmdHandler.Register("add", cmd.AddCommand, "Add a song to the queue !add <youtube-link>")
	print("registering commands\n")
	CmdHandler.Register("skip", cmd.SkipCommand, "Skip")
	print("registering commands\n")
	CmdHandler.Register("queue", cmd.QueueCommand, "Print queue???")
	print("registering commands\n")
	CmdHandler.Register("youtube", cmd.YoutubeCommand, "???")
	print("registering commands\n")
}
