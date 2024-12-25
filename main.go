package main

//#cgo LDFLAGS: -L . -lntgcalls -Wl,-rpath=./
import "C"
import (
	"fmt"

	"main/ntgcalls"

	"flag"

	"github.com/amarnathcjd/gogram/telegram"
	dotenv "github.com/joho/godotenv"
)

// media.Video = &VideoDescription{
// 		InputMode: InputModeShell,
// 		Input:     fmt.Sprintf("ffmpeg -i %s -loglevel panic -f rawvideo -r 24 -pix_fmt yuv420p -vf scale=1280:720 pipe:1", video),
// 		Width:     1280,
// 		Height:    720,
// 		Fps:       24,
// }

func main() {
	dotenv.Load(".env")
	file := flag.String("file", "out.mp3", "audio file path")
	group := flag.String("group", "@rosexchat", "group to join")
	flag.Parse()

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   AskInputOrEnv[int32]("API_KEY"),
		AppHash: AskInputOrEnv[string]("API_HASH"),
		// StringSession: AskInputOrEnv[string]("StringSession"),
	})

	client.AuthPrompt()
	if err != nil {
		panic(err)
	}

	ntg := ntgcalls.NTgCalls()
	defer ntg.Free()

	joinGroupCall(ntg, client, *group, *file)
	client.Idle()
}

func joinGroupCall(ntg *ntgcalls.Client, client *telegram.Client, username string, url string) {
	me, _ := client.GetMe()
	rawChannel, _ := client.ResolveUsername(username)
	channel := rawChannel.(*telegram.Channel)
	jsonParams, _ := ntg.CreateCall(channel.ID, ntgcalls.MediaDescription{
		Microphone: &ntgcalls.AudioDescription{
			MediaSource:  ntgcalls.MediaSourceShell, // ntgcalls.MediaSourceFile
			SampleRate:   128000,                    // 96000
			ChannelCount: 2,
			Input:        fmt.Sprintf("ffmpeg -i %s -loglevel panic -f s16le -ac 2 -ar 128k pipe:1", url), // './file.s16le'
		},
	})
	call, err := client.GetGroupCall(channel.ID)
	if err != nil {
		panic(err)
	}

	callResRaw, _ := client.PhoneJoinGroupCall(
		&telegram.PhoneJoinGroupCallParams{
			Muted:        false,
			VideoStopped: true,
			Call:         call,
			Params: &telegram.DataJson{
				Data: jsonParams,
			},
			JoinAs: &telegram.InputPeerUser{
				UserID:     me.ID,
				AccessHash: me.AccessHash,
			},
		},
	)
	callRes := callResRaw.(*telegram.UpdatesObj)
	for _, update := range callRes.Updates {
		switch update := update.(type) {
		case *telegram.UpdateGroupCallConnection:
			phoneCall := update
			_ = ntg.Connect(channel.ID, phoneCall.Params.Data, false)
		}
	}
}

func outgoingCall(client *ntgcalls.Client, mtproto *telegram.Client, username string) {
	var inputCall *telegram.InputPhoneCall

	rawUser, _ := mtproto.ResolveUsername(username)
	user := rawUser.(*telegram.UserObj)
	dhConfigRaw, _ := mtproto.MessagesGetDhConfig(0, 256)
	dhConfig := dhConfigRaw.(*telegram.MessagesDhConfigObj)
	_ = client.CreateP2PCall(user.ID, ntgcalls.MediaDescription{
		Microphone: &ntgcalls.AudioDescription{
			MediaSource:  ntgcalls.MediaSourceShell,
			SampleRate:   96000,
			ChannelCount: 2,
			Input:        "ffmpeg -reconnect 1 -reconnect_at_eof 1 -reconnect_streamed 1 -reconnect_delay_max 2 -i https://docs.evostream.com/sample_content/assets/sintel1m720p.mp4 -f s16le -ac 2 -ar 96k -v quiet pipe:1",
		},
	})
	gAHash, _ := client.InitExchange(user.ID, ntgcalls.DhConfig{
		G:      dhConfig.G,
		P:      dhConfig.P,
		Random: dhConfig.Random,
	}, nil)
	protocolRaw := ntgcalls.GetProtocol()
	protocol := &telegram.PhoneCallProtocol{
		UdpP2P:          protocolRaw.UdpP2P,
		UdpReflector:    protocolRaw.UdpReflector,
		MinLayer:        protocolRaw.MinLayer,
		MaxLayer:        protocolRaw.MaxLayer,
		LibraryVersions: protocolRaw.Versions,
	}
	_, _ = mtproto.PhoneRequestCall(
		&telegram.PhoneRequestCallParams{
			Protocol: protocol,
			UserID:   &telegram.InputUserObj{UserID: user.ID, AccessHash: user.AccessHash},
			GAHash:   gAHash,
			RandomID: int32(telegram.GenRandInt()),
		},
	)

	mtproto.AddRawHandler(&telegram.UpdatePhoneCall{}, func(m telegram.Update, c *telegram.Client) error {
		phoneCall := m.(*telegram.UpdatePhoneCall).PhoneCall
		switch phoneCall.(type) {
		case *telegram.PhoneCallAccepted:
			call := phoneCall.(*telegram.PhoneCallAccepted)
			res, _ := client.ExchangeKeys(user.ID, call.GB, 0)
			inputCall = &telegram.InputPhoneCall{
				ID:         call.ID,
				AccessHash: call.AccessHash,
			}
			client.OnSignal(func(chatId int64, signal []byte) {
				_, _ = mtproto.PhoneSendSignalingData(inputCall, signal)
			})
			callConfirmRes, _ := mtproto.PhoneConfirmCall(
				inputCall,
				res.GAOrB,
				res.KeyFingerprint,
				protocol,
			)
			callRes := callConfirmRes.PhoneCall.(*telegram.PhoneCallObj)
			rtcServers := make([]ntgcalls.RTCServer, len(callRes.Connections))
			for i, connection := range callRes.Connections {
				switch connection := connection.(type) {
				case *telegram.PhoneConnectionWebrtc:
					rtcServer := connection
					rtcServers[i] = ntgcalls.RTCServer{
						ID:       rtcServer.ID,
						Ipv4:     rtcServer.Ip,
						Ipv6:     rtcServer.Ipv6,
						Username: rtcServer.Username,
						Password: rtcServer.Password,
						Port:     rtcServer.Port,
						Turn:     rtcServer.Turn,
						Stun:     rtcServer.Stun,
					}
				case *telegram.PhoneConnectionObj:
					phoneServer := connection
					rtcServers[i] = ntgcalls.RTCServer{
						ID:      phoneServer.ID,
						Ipv4:    phoneServer.Ip,
						Ipv6:    phoneServer.Ipv6,
						Port:    phoneServer.Port,
						Turn:    true,
						Tcp:     phoneServer.Tcp,
						PeerTag: phoneServer.PeerTag,
					}
				}
			}
			_ = client.ConnectP2P(user.ID, rtcServers, callRes.Protocol.LibraryVersions, callRes.P2PAllowed)
		}
		return nil
	})

	mtproto.AddRawHandler(&telegram.UpdatePhoneCallSignalingData{}, func(m telegram.Update, c *telegram.Client) error {
		signalingData := m.(*telegram.UpdatePhoneCallSignalingData).Data
		_ = client.SendSignalingData(user.ID, signalingData)
		return nil
	})
}
