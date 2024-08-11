<h1>Telegram Group Calls Example</h1>

<p>This is a simple example of how to use the <a href="https://github.com/amarnathcjd/gogram">Gogram Library</a> to stream audio from a file to a Telegram group call.</p>

<h2>🚀 Pre-requisites</h2>
<ul>
    <li><strong>Go 1.18</strong> or higher</li>
    <li><strong>GCC compiler</strong> (for cgo)</li>
    <li><strong>FFmpeg</strong> (for audio pipeline)</li>
</ul>

<h2>🔧 Installation</h2>

```bash
git clone https://github.com/amarnathcjd/tgcalls.git
cd tgcalls

go env -w CGO_ENABLED=1
go mod tidy

go run . -file <path-to-audio-file> -group <group-username>
```

<h2>📚 NTgCalls</h2>
<p>This example uses the <a href="https://github.com/pytgcalls/ntgcalls">NTgCalls</a> library to interact with the Telegram WebRTC API.</p>
