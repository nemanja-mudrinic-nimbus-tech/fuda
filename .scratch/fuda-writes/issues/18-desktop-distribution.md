# Shipping the Wails desktop app at $0

Type: research
Status: resolved
Blocked by: 

## Question

What does it cost to ship a Wails v3 app on macOS, Windows and Linux without paid code signing? What do users see (Gatekeeper, SmartScreen) and how do they get past it? Free signing options (e.g. SignPath for open source, Azure Trusted Signing price)? How does Wails v3 auto-update work, and can it use GitHub Releases for free?

## Answer

Unsigned builds work everywhere, but cost user friction. macOS 15 dropped the right-click Open bypass: users allow the app in System Settings (or run `xattr`); only Apple's $99/year program avoids it. Windows shows SmartScreen ("More info" > "Run anyway"). Free Windows signing: SignPath Foundation, for OSI-licensed open source with CI builds. Azure Artifact Signing costs $9.99/month, and individuals must be in the US or Canada. Wails v3 (still beta) has a built-in updater that can read GitHub Releases, which are free with no bandwidth limit.

Details: branch `research/desktop-distribution`, file `.scratch/fuda-writes/research/18-desktop-distribution.md`.
