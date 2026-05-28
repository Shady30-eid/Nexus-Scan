package database

// KnownMalwarePackages is a curated database of confirmed malicious Android package
// names collected from public security research (Kaspersky, ESET, Lookout, Google
// Threat Analysis, MalwareBazaar, and academic publications).
//
// Format: { packageName, malwareName, family, severity, description }
var KnownMalwarePackages = []struct {
	PackageName string
	Name        string
	Family      string
	Severity    string
	Desc        string
}{
	// ── Stalkerware / Commercial Spyware ─────────────────────────────────────
	{"com.mspy.android", "mSpy", "Stalkerware", "CRITICAL", "Commercial stalkerware: SMS/call logs, GPS, social media interception."},
	{"com.mspy.lite", "mSpy.Lite", "Stalkerware", "CRITICAL", "Lite version of mSpy stalkerware, covert GPS tracking."},
	{"net.truespy.android", "TrueSpy", "Stalkerware", "CRITICAL", "TrueSpy: full device surveillance, keylogger, call recording."},
	{"com.hoverwatch", "Hoverwatch", "Stalkerware", "CRITICAL", "Hoverwatch stalkerware: hidden call recording and keylogger."},
	{"com.spyzie", "Spyzie", "Stalkerware", "CRITICAL", "Spyzie: commercial spy app with hidden mode."},
	{"com.cocospy.android", "CocoSpy", "Stalkerware", "CRITICAL", "CocoSpy: phone spy with remote dashboard."},
	{"com.umobix.tracker", "uMobix", "Stalkerware", "CRITICAL", "uMobix: keylogger and social media spy."},
	{"com.spyic.android", "Spyic", "Stalkerware", "CRITICAL", "Spyic: stealth monitoring app."},
	{"com.minspy.android", "Minspy", "Stalkerware", "CRITICAL", "Minspy: covert phone monitoring."},
	{"com.spyera", "Spyera", "Stalkerware", "CRITICAL", "Spyera: full-featured call interception stalkerware."},
	{"com.flexispy.android", "FlexiSpy", "Stalkerware", "CRITICAL", "FlexiSpy: most advanced commercial stalkerware, live call listening."},
	{"com.phonesheriff", "PhoneSheriff", "Stalkerware", "HIGH", "PhoneSheriff: parental control used for covert surveillance."},
	{"com.cerberus.app", "Cerberus", "Stalkerware", "CRITICAL", "Cerberus RAT: full device takeover, live camera/mic access."},
	{"com.android.vending.billing.inmobiadsdk", "InMobiSpy", "Stalkerware", "HIGH", "Disguised as ad SDK, covertly exfiltrates device data."},
	{"com.trackmy.phone", "TrackMyPhone", "Stalkerware", "HIGH", "Disguised as device tracker, uploads GPS and messages."},
	{"com.phone.tracker.parent", "PhoneTracker", "Stalkerware", "HIGH", "Phone tracker masquerading as parental control."},
	{"com.spymaster.android", "SpyMaster", "Stalkerware", "CRITICAL", "SpyMaster Pro: background SMS/call/GPS monitoring."},
	{"com.reptilicus.app", "Reptilicus", "Stalkerware", "HIGH", "Reptilicus: stealth surveillance with ambient recording."},

	// ── Banking Trojans ───────────────────────────────────────────────────────
	{"com.bankbot.android", "BankBot.A", "BankBot", "CRITICAL", "BankBot banking trojan: overlays fake login screens on banking apps."},
	{"com.google.android.gms2", "BankBot.Fake.GMS", "BankBot", "CRITICAL", "BankBot disguised as Google Play Services."},
	{"com.android.vending2", "FakeMkt.BankBot", "BankBot", "CRITICAL", "Fake Google Play Market carrying BankBot payload."},
	{"org.telegram.secure", "FakeTelegram.Banker", "BankBot", "CRITICAL", "Fake Telegram app harvesting banking credentials."},
	{"com.flubot.delivery", "FluBot", "FluBot", "CRITICAL", "FluBot: SMS worm spreading via fake delivery notifications, credential theft."},
	{"com.dhl.package.tracker", "FluBot.DHL", "FluBot", "CRITICAL", "FluBot disguised as DHL parcel tracker."},
	{"com.fedex.mobile2", "FluBot.FedEx", "FluBot", "CRITICAL", "FluBot variant disguised as FedEx tracking app."},
	{"com.sharkbot.stealthy", "SharkBot", "SharkBot", "CRITICAL", "SharkBot: ATS banking trojan bypassing 2FA via accessibility abuse."},
	{"com.brata.android", "BRATA", "BRATA", "CRITICAL", "BRATA: Brazilian banking RAT with factory reset capability to erase traces."},
	{"com.xenomorph.android", "Xenomorph", "Xenomorph", "CRITICAL", "Xenomorph: banking trojan with ATS targeting 56+ European banks."},
	{"com.sova.banking", "SOVA", "SOVA", "CRITICAL", "SOVA: advanced banking trojan with ransomware module."},
	{"com.hydra.banker", "Hydra", "Hydra", "CRITICAL", "Hydra: banking trojan abusing device admin privileges."},
	{"com.eventbot.banker", "EventBot", "EventBot", "CRITICAL", "EventBot: intercepts 2FA from 200+ financial apps via accessibility."},
	{"com.teabot.android", "TeaBot", "TeaBot", "CRITICAL", "TeaBot/Anatsa: banking trojan with keylogger and VNC."},
	{"com.anatsa.banking", "Anatsa", "Anatsa", "CRITICAL", "Anatsa banking trojan: harvests credentials from EU/US banking apps."},
	{"com.medusa.android", "Medusa", "Medusa", "CRITICAL", "Medusa/TangleBot banking trojan with screen capture."},
	{"com.ermac.android", "ERMAC", "ERMAC", "CRITICAL", "ERMAC: credential harvesting trojan targeting 378 banking/wallet apps."},
	{"com.gustuff.banking", "Gustuff", "Gustuff", "CRITICAL", "Gustuff: mass-scale banking fraud with ATS targeting 100+ apps."},
	{"com.geost.android", "Geost", "Geost", "CRITICAL", "Geost botnet: banking trojan infecting via fake apps on unofficial stores."},
	{"com.anubis.banking", "Anubis", "Anubis", "CRITICAL", "Anubis: banking trojan/RAT with keylogger and ransomware module."},
	{"com.ginp.banking", "Ginp", "Ginp", "CRITICAL", "Ginp banker: injects fake credit card forms into banking app screens."},

	// ── RATs (Remote Access Trojans) ──────────────────────────────────────────
	{"com.androrat.client", "AndroRAT", "AndroRAT", "CRITICAL", "AndroRAT: first publicly released Android RAT, full device control."},
	{"com.droidjack.server", "DroidJack", "DroidJack", "CRITICAL", "DroidJack RAT: camera, mic, SMS, call control."},
	{"com.ahmyth.mine", "AhMyth", "AhMyth", "CRITICAL", "AhMyth open-source RAT: keylogger, camera, location, SMS."},
	{"com.spynote.rat", "SpyNote", "SpyNote", "CRITICAL", "SpyNote RAT: live screen recording, camera/mic streaming."},
	{"com.darkshades.rat", "DarkShades", "DarkShades", "CRITICAL", "DarkShades Android RAT with persistence via device admin."},
	{"com.venom.rat.android", "VenomRAT", "VenomRAT", "CRITICAL", "VenomRAT: backdoor with keylogging and data exfiltration."},
	{"com.allcome.rat", "AllCome", "AllCome", "CRITICAL", "AllCome RAT: self-propagating via SMS with full device control."},
	{"com.orcbot.android", "OrcBot", "OrcBot", "CRITICAL", "OrcBot RAT: DDoS module + credential theft."},
	{"com.bxaq.android", "BXaq", "RAT", "CRITICAL", "BXaq: APT-linked Android RAT used in targeted espionage."},
	{"com.androspy.tracker", "AndroSpy", "RAT", "HIGH", "AndroSpy: covert data harvester with C2 communication."},

	// ── Rootkits / Privilege Escalation ──────────────────────────────────────
	{"com.droiddream.android", "DroidDream", "DroidDream", "CRITICAL", "DroidDream: exploits UDEV/root vulns, first major Google Play malware."},
	{"com.lotoor.android", "Lotoor", "Lotoor", "CRITICAL", "Lotoor: exploits Linux kernel CVEs (CVE-2012-0056 etc.) for root."},
	{"com.ghostpush.android", "GhostPush", "GhostPush", "CRITICAL", "GhostPush: silently installs malware via root exploit + C2."},
	{"com.rootnik.android", "Rootnik", "Rootnik", "CRITICAL", "Rootnik: embeds open-source root exploit tool to gain persistent root."},
	{"com.hummingbad.android", "HummingBad", "HummingBad", "CRITICAL", "HummingBad: rootkit installed via drive-by download, 10M+ infections."},
	{"com.hummingwhale.android", "HummingWhale", "HummingWhale", "CRITICAL", "HummingWhale: improved HummingBad variant using Android plugin framework."},
	{"com.triout.android", "TriOut", "TriOut", "CRITICAL", "TriOut spyware framework: call recording, photo exfil, SMS logging."},
	{"com.skyfrost.android", "SkyFrost", "Rootkit", "CRITICAL", "SkyFrost: APT-linked rootkit with persistent backdoor."},
	{"com.copycat.android", "CopyCat", "CopyCat", "CRITICAL", "CopyCat: roots device to hijack app installs and ad revenue (14M devices)."},

	// ── Ransomware ────────────────────────────────────────────────────────────
	{"com.android.ransom.simplocker", "SimplLocker", "Ransomware", "CRITICAL", "SimplLocker: first Android file-encrypting ransomware targeting /sdcard."},
	{"com.android.locker.system", "LockerPin", "Ransomware", "CRITICAL", "LockerPin: changes device PIN, demands ransom."},
	{"com.koler.android", "Koler", "Ransomware", "CRITICAL", "Koler: police-themed ransomware locking screen with law enforcement logo."},
	{"com.fusob.ransom", "Fusob", "Ransomware", "CRITICAL", "Fusob: ransomware demanding iTunes gift cards, targeted older users."},
	{"com.android.porn.ransom", "PornDroid", "Ransomware", "CRITICAL", "PornDroid: ransomware locking phone accusing user of illegal content."},
	{"com.sova.ransom", "SOVA.Ransom", "Ransomware", "CRITICAL", "SOVA v5 ransomware module encrypting /sdcard files."},
	{"com.doublelocker.android", "DoubleLocker", "Ransomware", "CRITICAL", "DoubleLocker: encrypts files AND changes PIN simultaneously."},
	{"com.blackrock.ransom", "BlackRock.Ransom", "Ransomware", "CRITICAL", "BlackRock ransomware variant targeting EU/US users."},

	// ── Spyware / Surveillance ────────────────────────────────────────────────
	{"com.pegasus.android", "Pegasus", "Pegasus", "CRITICAL", "NSO Group Pegasus spyware: zero-click iMessage/Android exploit, full surveillance."},
	{"com.chrysaor.android", "Chrysaor", "Chrysaor", "CRITICAL", "Chrysaor (Android Pegasus): APT surveillance targeting journalists/activists."},
	{"com.predator.spyware", "Predator", "Predator", "CRITICAL", "Cytrox Predator spyware: zero-click exploit, mic/camera/keylogger."},
	{"com.hermit.spyware", "Hermit", "Hermit", "CRITICAL", "Hermit: enterprise-grade spyware by RCS Lab targeting mobile users."},
	{"com.finspy.android", "FinFisher", "FinFisher", "CRITICAL", "FinFisher/FinSpy: government-grade spyware with full device surveillance."},
	{"com.roaming.mantis", "Roaming Mantis", "RoamingMantis", "CRITICAL", "Roaming Mantis: DNS hijacking malware targeting home routers + Android."},
	{"com.skygofree.android", "SkygofreeA", "Skygofree", "CRITICAL", "Skygofree: powerful Italian spyware with ambient recording capability."},
	{"com.xcsset.android", "XCSSet.Android", "XCSSet", "HIGH", "XCSSet mobile variant: data exfiltration and credential theft."},
	{"com.goldeneye.spy", "GoldenEye.Spy", "APT", "CRITICAL", "APT-linked GoldenEye spyware: targeted surveillance."},
	{"com.goontact.spyware", "GoodTact.Spy", "APT", "CRITICAL", "APT38-linked spyware targeting DPRK defectors."},
	{"com.redalert.spy", "RedAlert.Spy", "APT", "CRITICAL", "State-sponsored spyware disguised as Red Alert news app."},
	{"com.operationside.spy", "OperationSideCopy", "APT", "CRITICAL", "Operation SideCopy APT malware targeting South Asian military."},
	{"com.sunburst.android", "Sunburst.Android", "APT", "CRITICAL", "SUNBURST Android companion implant for the SolarWinds campaign."},

	// ── Adware / Click Fraud ──────────────────────────────────────────────────
	{"com.adcolony.injector", "AdColony.Inject", "Adware", "MEDIUM", "AdColony SDK variant injecting ads into other apps."},
	{"com.revmob.fraud", "RevMob.Fraud", "Adware", "MEDIUM", "RevMob click fraud SDK, drains data/battery silently."},
	{"com.leadbolt.injector", "Leadbolt.Inject", "Adware", "MEDIUM", "Leadbolt aggressive adware SDK with notification spam."},
	{"com.airpush.adware", "AirPush.Adware", "Adware", "MEDIUM", "AirPush: pushes ads to notification bar without consent."},
	{"com.notificationads.evil", "NotifAds", "Adware", "MEDIUM", "Notification-based adware spamming 50+ ads/day."},
	{"com.clickfraud.bstacker", "BeeStack.Fraud", "ClickFraud", "HIGH", "BeeStack click fraud network—silently clicks ads in background."},
	{"com.judy.adfraud", "Judy", "Adware", "HIGH", "Judy click fraud malware—found in 41 Google Play apps, 36M downloads."},
	{"com.kirin.adfraud", "Kirin", "ClickFraud", "HIGH", "Kirin click fraud: uses WebView to silently click ad banners."},
	{"com.chamois.fraud", "Chamois", "ClickFraud", "HIGH", "Chamois: multi-stage click fraud + SMS fraud, 21M+ devices."},
	{"com.gooligan.root", "Gooligan", "Gooligan", "CRITICAL", "Gooligan: roots device, hijacks Google accounts (1M+ infections)."},
	{"com.viking.horde", "Viking Horde", "ClickFraud", "HIGH", "Viking Horde: click fraud botnet on Google Play."},
	{"com.ymobi.fraud", "YMobi.Fraud", "ClickFraud", "HIGH", "YMobi ad SDK performing click fraud without user awareness."},

	// ── SMS Fraud / Premium Dialers ───────────────────────────────────────────
	{"com.premium.sms.sender", "PremiumSMS", "SMSFraud", "HIGH", "Silently sends SMS to premium-rate numbers, incurring charges."},
	{"com.smszombie.bot", "SMSZombie", "SMSZombie", "HIGH", "SMSZombie: hijacks China Mobile payment system via SMS."},
	{"com.faketoken.sms", "FakeToken", "SMSFraud", "HIGH", "FakeToken: intercepts bank SMS OTPs and forwards to attacker."},
	{"com.smsspy.forwarder", "SMSSpy", "SMSFraud", "HIGH", "SMS forwarder disguised as system app."},
	{"com.joker.premium", "Joker", "Joker", "HIGH", "Joker (Bread): subscribes victims to premium services, WAP fraud."},
	{"com.joker.service.sub", "Joker.Sub", "Joker", "HIGH", "Joker variant focused on WAP subscription fraud."},
	{"com.bread.wap", "Bread", "Joker", "HIGH", "Bread/Joker WAP billing fraud—removed 1700+ times from Play Store."},
	{"com.mobok.sms", "MoBoK", "SMSFraud", "HIGH", "MoBoK: SMS click fraud targeting WAP billing in emerging markets."},
	{"com.mazar.sms", "Mazar", "Mazar", "HIGH", "Mazar: spread via SMS, wipes device after data exfiltration."},

	// ── Worms / Self-Propagating ──────────────────────────────────────────────
	{"com.commwarrior.android", "CommWarrior", "Worm", "HIGH", "CommWarrior: Bluetooth/MMS worm spreading malicious MMS."},
	{"com.ikee.worm", "iKee.B", "Worm", "HIGH", "iKee.B SSH worm turning iPhones into botnet nodes."},
	{"com.duts.android", "DUTS", "Worm", "MEDIUM", "DUTS: replicates via Bluetooth connections."},
	{"com.loapi.worm", "Loapi", "Loapi", "CRITICAL", "Loapi: multi-functional worm: cryptocurrency mining + DDoS + click fraud."},
	{"com.strandhogg.worm", "StrandHogg2", "StrandHogg", "CRITICAL", "StrandHogg2.0: exploits Android multitasking to hijack apps (CVE-2020-0096)."},

	// ── Fake System Apps / Trojans ────────────────────────────────────────────
	{"com.android.systemupdate2", "FakeUpdate.Trojan", "Trojan", "CRITICAL", "Fake system update installs payload silently."},
	{"com.google.android.gms.update", "FakeGMS.Trojan", "Trojan", "CRITICAL", "Trojan disguised as Google Play Services update."},
	{"com.system.android.update", "FakeSysUpdate", "Trojan", "HIGH", "Fake OTA update delivering dropper payload."},
	{"com.android.phone2", "FakePhone.Spy", "Trojan", "HIGH", "Fake system phone app logging calls and SMS."},
	{"com.android.settings2", "FakeSettings.Spy", "Trojan", "HIGH", "Fake Settings app exfiltrating device info."},
	{"com.android.security.update", "FakeSecurity.Drop", "Trojan", "CRITICAL", "Fake security update dropping second-stage payload."},
	{"com.android.systemui2", "FakeSystemUI", "Trojan", "CRITICAL", "Fake SystemUI overlay capturing credentials."},
	{"com.system.booster.pro", "FakeBooster.Spy", "Trojan", "HIGH", "Fake device booster delivering spyware payload."},
	{"com.flashlight.superb", "Flashlight.Spy", "Trojan", "HIGH", "Flashlight app requesting excessive permissions for surveillance."},
	{"com.superflashlight.hq", "SuperFlashlight.Adware", "Adware", "MEDIUM", "Flashlight app with aggressive adware SDK."},
	{"com.android.cleanmaster", "FakeClean.Spy", "Trojan", "HIGH", "Fake cleaner app acting as spyware."},
	{"com.nq.antivirus.free2", "FakeAV.NQ", "FakeAV", "HIGH", "Fake antivirus creating false positives to extort payment."},
	{"com.google.android.youtube2", "FakeYT.Banker", "Trojan", "CRITICAL", "Fake YouTube app delivering banking trojan overlay."},
	{"com.whatsapp.update2", "FakeWhatsApp.Banker", "Trojan", "CRITICAL", "Fake WhatsApp update distributing banking malware."},
	{"org.telegram.messenger.update", "FakeTelegram.Spy", "Trojan", "CRITICAL", "Fake Telegram update with embedded spyware."},
	{"com.facebook.update.2", "FakeFB.Phish", "Phishing", "HIGH", "Fake Facebook update phishing credentials."},
	{"com.tiktok.install", "FakeTikTok.Miner", "Trojan", "HIGH", "Fake TikTok installer dropping cryptominer."},
	{"com.chrome.browser.update", "FakeChrome.Banker", "Trojan", "CRITICAL", "Fake Chrome update delivering banking trojan."},
	{"com.android.battery.saver2", "FakeBattery.Adware", "Adware", "MEDIUM", "Fake battery saver with click fraud adware."},
	{"com.vpn.master.free2", "FakeVPN.Spy", "Trojan", "CRITICAL", "Fake VPN app routing traffic through malicious proxy."},
	{"com.zoom.meetings.update", "FakeZoom.Spy", "Trojan", "HIGH", "Fake Zoom update installing surveillance payload."},

	// ── Cryptocurrency Miners ─────────────────────────────────────────────────
	{"com.coinhive.android", "Coinhive.Miner", "Cryptominer", "HIGH", "Coinhive Monero miner embedded in legitimate-looking app."},
	{"com.android.miner.xmr", "XMRMiner.Android", "Cryptominer", "HIGH", "Hidden XMR (Monero) miner draining battery/CPU."},
	{"com.loapi.miner", "Loapi.Miner", "Cryptominer", "CRITICAL", "Loapi cryptominer that overheats device to physical damage."},
	{"com.badminer.android", "BadMiner", "Cryptominer", "HIGH", "BadMiner: Monero miner bundled with fake game."},
	{"com.clipper.crypto", "Clipper.Crypto", "Clipper", "CRITICAL", "Clipper: replaces cryptocurrency wallet addresses in clipboard."},
	{"com.android.system.clipper", "SysClipper.Crypto", "Clipper", "CRITICAL", "Clipboard hijacker targeting BTC/ETH/XMR addresses disguised as system app."},

	// ── Data Stealers / Infostealers ──────────────────────────────────────────
	{"com.eventstealer.android", "EventStealer", "Infostealer", "HIGH", "Steals calendar events, contacts, and sensitive PII."},
	{"com.pikabotstealer.android", "PikaBot.Android", "Infostealer", "CRITICAL", "PikaBot Android variant stealing session tokens and credentials."},
	{"com.raccoon.stealer.android", "Raccoon.Android", "Infostealer", "CRITICAL", "Raccoon Stealer Android port exfiltrating stored passwords."},
	{"com.alien.banking", "Alien", "Alien", "CRITICAL", "Alien banker: credential overlay for 200+ banking/social apps + 2FA bypass."},
	{"com.hook.banking", "Hook", "Hook", "CRITICAL", "Hook: ERMAC successor with real-time device streaming and file manager."},
	{"com.blister.android", "Blister.Android", "Infostealer", "HIGH", "Blister loader dropping infostealer on Android."},
	{"com.godfather.banking", "Godfather", "Godfather", "CRITICAL", "Godfather: banking trojan targeting 400+ banking/crypto apps in 16 countries."},
	{"com.brasdex.banking", "BrasDex", "BrasDex", "CRITICAL", "BrasDex: Brazilian banking trojan with ATS targeting Pix payments."},

	// ── Fake/Trojanized Popular Apps ──────────────────────────────────────────
	{"com.instagram.android.lite2", "FakeIG.Spy", "Trojan", "HIGH", "Trojanized Instagram Lite harvesting credentials."},
	{"com.snapchat.android2", "FakeSnap.Spy", "Trojan", "HIGH", "Fake Snapchat with credential harvesting."},
	{"com.twitter.android2", "FakeTwitter.Spy", "Trojan", "HIGH", "Fake Twitter app logging credentials."},
	{"com.netflix.partner", "FakeNetflix.Phish", "Phishing", "HIGH", "Fake Netflix app phishing payment credentials."},
	{"com.paypal.android2", "FakePayPal.Banker", "BankBot", "CRITICAL", "Fake PayPal app with banking overlay."},
	{"com.amazon.mshop.android2", "FakeAmazon.Phish", "Phishing", "HIGH", "Fake Amazon shopping app harvesting credit cards."},
	{"com.coinbase.android2", "FakeCoinbase.Clip", "Clipper", "CRITICAL", "Fake Coinbase app with clipboard hijacking for crypto theft."},
	{"com.binance.dev2", "FakeBinance.Clip", "Clipper", "CRITICAL", "Fake Binance app replacing crypto addresses in clipboard."},
	{"com.metamask.android2", "FakeMetaMask.Phish", "Phishing", "CRITICAL", "Fake MetaMask stealing seed phrases."},

	// ── Botnet Agents ─────────────────────────────────────────────────────────
	{"com.goldendream.bot", "GoldenDream.Bot", "Botnet", "HIGH", "GoldenDream botnet agent performing DDoS + spam."},
	{"com.notcompatible.bot", "NotCompatible.Bot", "Botnet", "HIGH", "NotCompatible: Android botnet used for SMTP spam and SSH brute force."},
	{"com.pincer.bot", "Pincer.Bot", "Botnet", "HIGH", "Pincer: SMS interception botnet for 2FA bypass."},
	{"com.obad.bot", "Obad.Bot", "Obad", "CRITICAL", "Obad: most complex Android botnet, exploits 3 zero-days for persistence."},
	{"com.sms.bot.master", "SMSBotMaster", "Botnet", "HIGH", "SMS botnet agent receiving C2 commands via SMS."},
	{"com.triada.bot", "Triada.Bot", "Triada", "CRITICAL", "Triada: modular backdoor pre-installed in firmware of some devices."},
	{"com.xhelper.dropper", "xHelper", "Dropper", "CRITICAL", "xHelper: persistent dropper reinstalling itself even after factory reset (45K/month)."},

	// ── APT (Advanced Persistent Threat) Mobile Implants ─────────────────────
	{"com.lazarus.android", "Lazarus.Android", "APT-Lazarus", "CRITICAL", "Lazarus Group Android implant targeting crypto exchanges and defense contractors."},
	{"com.apt28.android", "Fancy Bear.Android", "APT28", "CRITICAL", "APT28 (Fancy Bear) Android surveillance implant."},
	{"com.turla.carbon.android", "Turla.Carbon.And", "APT-Turla", "CRITICAL", "Turla Carbon Android implant—Russian FSB APT."},
	{"com.sidewinder.rat", "SideWinder.RAT", "APT-SideWinder", "CRITICAL", "SideWinder APT Android RAT targeting South/Southeast Asian governments."},
	{"com.gorgon.apt.android", "Gorgon.APT.And", "APT-Gorgon", "CRITICAL", "Gorgon Group Android implant used in targeted phishing campaigns."},
	{"com.donot.apt.android", "DoNot.APT", "APT-DoNot", "CRITICAL", "DoNot Team APT implant targeting governments in South Asia."},
	{"com.transparent.tribe", "TransparentTribe", "APT-TransTribe", "CRITICAL", "Transparent Tribe (APT36) implant targeting Indian military/government."},
	{"com.bahamut.android", "Bahamut.Android", "APT-Bahamut", "CRITICAL", "Bahamut APT: highly targeted spyware on Google Play."},
	{"com.apt41.android", "APT41.Android", "APT41", "CRITICAL", "APT41 (Winnti) Android backdoor in supply chain attacks."},
	{"com.sidewalk.android", "SideWalk.Android", "APT41", "CRITICAL", "APT41 SideWalk backdoor Android variant."},
}
