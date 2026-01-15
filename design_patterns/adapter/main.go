package main

import "fmt"

// 适配器模式：将一个类的接口转换成客户希望的另一个接口

// 目标接口（客户端期望的接口）
type MediaPlayer interface {
	Play(audioType string, fileName string)
}

// 高级媒体播放器接口
type AdvancedMediaPlayer interface {
	PlayVlc(fileName string)
	PlayMp4(fileName string)
}

// VLC播放器
type VlcPlayer struct{}

func (v *VlcPlayer) PlayVlc(fileName string) {
	fmt.Printf("播放 VLC 文件: %s\n", fileName)
}

func (v *VlcPlayer) PlayMp4(fileName string) {
	// VLC不支持MP4
}

// MP4播放器
type Mp4Player struct{}

func (m *Mp4Player) PlayVlc(fileName string) {
	// MP4播放器不支持VLC
}

func (m *Mp4Player) PlayMp4(fileName string) {
	fmt.Printf("播放 MP4 文件: %s\n", fileName)
}

// 媒体适配器
type MediaAdapter struct {
	advancedPlayer AdvancedMediaPlayer
}

func NewMediaAdapter(audioType string) *MediaAdapter {
	if audioType == "vlc" {
		return &MediaAdapter{advancedPlayer: &VlcPlayer{}}
	} else if audioType == "mp4" {
		return &MediaAdapter{advancedPlayer: &Mp4Player{}}
	}
	return nil
}

func (m *MediaAdapter) Play(audioType string, fileName string) {
	if audioType == "vlc" {
		m.advancedPlayer.PlayVlc(fileName)
	} else if audioType == "mp4" {
		m.advancedPlayer.PlayMp4(fileName)
	}
}

// 音频播放器（支持MP3，通过适配器支持VLC和MP4）
type AudioPlayer struct {
	mediaAdapter *MediaAdapter
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{}
}

func (a *AudioPlayer) Play(audioType string, fileName string) {
	if audioType == "mp3" {
		fmt.Printf("播放 MP3 文件: %s\n", fileName)
	} else if audioType == "vlc" || audioType == "mp4" {
		a.mediaAdapter = NewMediaAdapter(audioType)
		a.mediaAdapter.Play(audioType, fileName)
	} else {
		fmt.Printf("不支持的音频格式: %s\n", audioType)
	}
}

func main() {
	player := NewAudioPlayer()

	player.Play("mp3", "song.mp3")
	player.Play("mp4", "video.mp4")
	player.Play("vlc", "movie.vlc")
	player.Play("avi", "movie.avi")
}
