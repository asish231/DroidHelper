package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type device struct {
	Serial string
	State  string
	Model  string
	Raw    string
}

type mdnsService struct {
	Name     string
	Service  string
	Endpoint string
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("DroidHelper CLI")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("This helper checks scrcpy and adb, then guides you through USB or wireless mirroring.")
	fmt.Println()

	if err := ensureDependencies(reader); err != nil {
		exitWithError(err)
	}

	if err := startADB(); err != nil {
		exitWithError(err)
	}

	mode, err := chooseOption(reader, "Choose connection mode:", []string{
		"USB mirroring",
		"Wireless mirroring",
		"Exit",
	})
	if err != nil {
		exitWithError(err)
	}

	switch mode {
	case 0:
		if err := handleUSB(reader); err != nil {
			exitWithError(err)
		}
	case 1:
		if err := handleWireless(reader); err != nil {
			exitWithError(err)
		}
	default:
		fmt.Println("Exited.")
	}
}

func ensureDependencies(reader *bufio.Reader) error {
	if err := ensureCommand(reader, "scrcpy", "brew", "install", "scrcpy"); err != nil {
		return err
	}
	if err := ensureCommand(reader, "adb", "brew", "install", "--cask", "android-platform-tools"); err != nil {
		return err
	}
	return nil
}

func ensureCommand(reader *bufio.Reader, command string, installCmd ...string) error {
	if _, err := exec.LookPath(command); err == nil {
		fmt.Printf("%s is installed.\n", command)
		return nil
	}

	fmt.Printf("%s was not found.\n", command)
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew is not installed, so %s cannot be installed automatically", command)
	}

	install, err := askYesNo(reader, fmt.Sprintf("Do you want to install %s now?", command))
	if err != nil {
		return err
	}
	if !install {
		return fmt.Errorf("%s is required to continue", command)
	}

	fmt.Printf("Installing %s...\n", command)
	if err := runInteractive(installCmd[0], installCmd[1:]...); err != nil {
		return fmt.Errorf("failed to install %s: %w", command, err)
	}

	if _, err := exec.LookPath(command); err != nil {
		return fmt.Errorf("%s still was not found after installation", command)
	}

	fmt.Printf("%s installed successfully.\n", command)
	return nil
}

func handleUSB(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("USB mirroring")
	fmt.Println("-------------")
	fmt.Println("1. Connect your phone by USB.")
	fmt.Println("2. Enable USB debugging on the phone if needed.")
	fmt.Println()

	for {
		devices, err := listADBDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			fmt.Println("No adb devices were detected.")
			retry, err := askYesNo(reader, "Try again after connecting the phone?")
			if err != nil {
				return err
			}
			if retry {
				continue
			}
			return errors.New("no USB device available")
		}

		selected, err := chooseADBDevice(reader, devices)
		if err != nil {
			return err
		}

		fmt.Printf("I found your device: %s\n", selected.Serial)
		audioArgs, modeLabel, err := chooseAudioMode(reader)
		if err != nil {
			return err
		}

		args := []string{"-s", selected.Serial}
		args = append(args, audioArgs...)

		fmt.Printf("Launching scrcpy in %s mode...\n", modeLabel)
		return runInteractive("scrcpy", args...)
	}
}

func handleWireless(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("Wireless mirroring")
	fmt.Println("------------------")
	fmt.Println("Before continuing:")
	fmt.Println("1. Turn on Developer options > Wireless debugging on the phone.")
	fmt.Println("2. Make sure the phone and this Mac are on the same Wi-Fi network.")
	fmt.Println("3. Open 'Pair device with pairing code' on the phone.")
	fmt.Println()

	showLocalNetworkHints()

	mdns, _ := listMDNSServices()
	if len(mdns) > 0 {
		fmt.Println()
		fmt.Println("adb mDNS services already visible on the network:")
		printMDNSServices(mdns)
	}

	fmt.Println()
	phoneAddr, err := promptNonEmpty(reader, "Enter the phone IP address shown on the phone (or IP:pairingPort)")
	if err != nil {
		return err
	}
	ip, pairPort, err := parsePhoneAddress(phoneAddr)
	if err != nil {
		return err
	}

	connectServices, _ := listConnectServices(ip)
	if len(connectServices) > 0 {
		fmt.Println()
		fmt.Println("adb wireless connect services already visible for this phone:")
		printMDNSServices(connectServices)

		useExisting, err := askYesNo(reader, "Use one of these connect endpoints without pairing again?")
		if err != nil {
			return err
		}
		if useExisting {
			idx, err := chooseOption(reader, "Choose the connect endpoint to use:", endpoints(connectServices))
			if err != nil {
				return err
			}
			return connectAndLaunch(reader, connectServices[idx].Endpoint)
		}
	}

	if pairPort == "" {
		fmt.Println("Tip: the pairing port is temporary and changes often. It is not the same as the later connect port.")
		pairPort, err = promptNonEmpty(reader, "Enter the pairing port")
		if err != nil {
			return err
		}
	}
	pairCode, err := promptNonEmpty(reader, "Enter the pairing code")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Pinging %s...\n", ip)
	if err := runInteractive("ping", "-c", "1", ip); err != nil {
		return fmt.Errorf("the phone IP did not respond to ping: %w", err)
	}

	fmt.Printf("Checking whether %s:%s is reachable...\n", ip, pairPort)
	if err := runInteractive("nc", "-vz", ip, pairPort); err != nil {
		return fmt.Errorf("the pairing port did not respond: %w", err)
	}

	fmt.Println("Pairing with the phone...")
	if err := runInteractive("adb", "pair", net.JoinHostPort(ip, pairPort), pairCode); err != nil {
		return fmt.Errorf("adb pairing failed: %w", err)
	}

	connectEndpoint, err := discoverConnectEndpoint(reader, ip)
	if err != nil {
		return err
	}

	return connectAndLaunch(reader, connectEndpoint)
}

func connectAndLaunch(reader *bufio.Reader, connectEndpoint string) error {
	fmt.Printf("Connecting to %s...\n", connectEndpoint)
	if err := runInteractive("adb", "connect", connectEndpoint); err != nil {
		return fmt.Errorf("adb connect failed: %w", err)
	}

	audioArgs, modeLabel, err := chooseAudioMode(reader)
	if err != nil {
		return err
	}

	fmt.Printf("Launching scrcpy in %s mode...\n", modeLabel)
	args := []string{"-s", connectEndpoint}
	args = append(args, audioArgs...)
	return runInteractive("scrcpy", args...)
}

func parsePhoneAddress(input string) (string, string, error) {
	if parsed := net.ParseIP(input); parsed != nil {
		return input, "", nil
	}

	host, port, err := net.SplitHostPort(input)
	if err == nil && net.ParseIP(host) != nil && port != "" {
		return host, port, nil
	}

	return "", "", fmt.Errorf("invalid IP address: %s", input)
}

func discoverConnectEndpoint(reader *bufio.Reader, ip string) (string, error) {
	fmt.Println()
	fmt.Println("Looking for adb wireless connect services...")
	matches, _ := listConnectServices(ip)

	if len(matches) > 0 {
		fmt.Println("Found connect service(s):")
		printMDNSServices(matches)
		idx, err := chooseOption(reader, "Choose the connect endpoint to use:", endpoints(matches))
		if err != nil {
			return "", err
		}
		return matches[idx].Endpoint, nil
	}

	fmt.Println("No connect service was discovered automatically.")
	port, err := promptNonEmpty(reader, "Enter the connect port shown by adb mdns services or from a previous working session")
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(ip, port), nil
}

func listConnectServices(ip string) ([]mdnsService, error) {
	services, err := listMDNSServices()
	if err != nil {
		return nil, err
	}

	var matches []mdnsService
	for _, svc := range services {
		if svc.Service == "_adb-tls-connect._tcp" && strings.HasPrefix(svc.Endpoint, ip+":") {
			matches = append(matches, svc)
		}
	}
	return matches, nil
}

func chooseADBDevice(reader *bufio.Reader, devices []device) (device, error) {
	if len(devices) == 1 {
		return devices[0], nil
	}

	options := make([]string, 0, len(devices))
	for _, dev := range devices {
		label := dev.Serial
		if dev.Model != "" {
			label += " (" + dev.Model + ")"
		}
		if dev.State != "" {
			label += " [" + dev.State + "]"
		}
		options = append(options, label)
	}

	idx, err := chooseOption(reader, "Choose the device to use:", options)
	if err != nil {
		return device{}, err
	}
	return devices[idx], nil
}

func chooseAudioMode(reader *bufio.Reader) ([]string, string, error) {
	idx, err := chooseOption(reader, "Choose playback mode:", []string{
		"No audio",
		"Normal device audio on the computer",
		"Voice-call mode",
	})
	if err != nil {
		return nil, "", err
	}

	switch idx {
	case 0:
		return []string{"--no-audio"}, "no-audio", nil
	case 1:
		return []string{"--audio-source=output"}, "normal audio", nil
	case 2:
		return []string{"--audio-source=voice-call", "--require-audio"}, "voice-call", nil
	default:
		return nil, "", errors.New("invalid audio mode")
	}
}

func startADB() error {
	return runInteractive("adb", "start-server")
}

func listADBDevices() ([]device, error) {
	output, err := runOutput("adb", "devices", "-l")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var devices []device
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		serial := fields[0]
		state := fields[1]
		if state != "device" {
			continue
		}

		model := ""
		for _, field := range fields[2:] {
			if strings.HasPrefix(field, "model:") {
				model = strings.TrimPrefix(field, "model:")
			}
		}

		devices = append(devices, device{
			Serial: serial,
			State:  state,
			Model:  model,
			Raw:    line,
		})
	}

	return devices, nil
}

func listMDNSServices() ([]mdnsService, error) {
	output, err := runOutput("adb", "mdns", "services")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var services []mdnsService
	for _, line := range lines[1:] {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 3 {
			continue
		}
		services = append(services, mdnsService{
			Name:     fields[0],
			Service:  fields[1],
			Endpoint: fields[2],
		})
	}
	return services, nil
}

func printMDNSServices(services []mdnsService) {
	for _, svc := range services {
		fmt.Printf("- %s  %s  %s\n", svc.Name, svc.Service, svc.Endpoint)
	}
}

func endpoints(services []mdnsService) []string {
	list := make([]string, 0, len(services))
	for _, svc := range services {
		list = append(list, svc.Endpoint)
	}
	return list
}

func showLocalNetworkHints() {
	fmt.Println("Known local-network hints:")
	ips := arpIPs()
	if len(ips) == 0 {
		fmt.Println("- No local IPs found in the ARP cache yet.")
		fmt.Println("- That is fine. You can read the phone IP from the phone and enter it manually.")
		return
	}

	for _, ip := range ips {
		fmt.Printf("- %s\n", ip)
	}
	fmt.Println("If you do not know which IP is the phone, check the Wi-Fi details on the phone and pick that address.")
}

func arpIPs() []string {
	output, err := runOutput("arp", "-a")
	if err != nil {
		return nil
	}

	re := regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\)`)
	matches := re.FindAllStringSubmatch(output, -1)
	seen := map[string]bool{}
	var ips []string
	for _, match := range matches {
		ip := match[1]
		if seen[ip] {
			continue
		}
		seen[ip] = true
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	return ips
}

func chooseOption(reader *bufio.Reader, title string, options []string) (int, error) {
	for {
		fmt.Println()
		fmt.Println(title)
		for i, option := range options {
			fmt.Printf("%d. %s\n", i+1, option)
		}
		fmt.Print("> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			fmt.Println("Enter a number.")
			continue
		}

		var choice int
		if _, err := fmt.Sscanf(text, "%d", &choice); err != nil {
			fmt.Println("Enter a valid number.")
			continue
		}
		if choice < 1 || choice > len(options) {
			fmt.Println("Choose one of the listed options.")
			continue
		}
		return choice - 1, nil
	}
}

func askYesNo(reader *bufio.Reader, prompt string) (bool, error) {
	for {
		fmt.Printf("%s [y/n]: ", prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		text = strings.ToLower(strings.TrimSpace(text))
		switch text {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}

func promptNonEmpty(reader *bufio.Reader, prompt string) (string, error) {
	for {
		fmt.Printf("%s: ", prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text != "" {
			return text, nil
		}
		fmt.Println("This value cannot be empty.")
	}
}

func runInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func exitWithError(err error) {
	fmt.Println()
	fmt.Println("Error:", err)
	fmt.Println()
	if runtime.GOOS == "windows" {
		fmt.Println("Press Enter to close...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(1)
}
