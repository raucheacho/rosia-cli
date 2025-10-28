# Homebrew Formula Template for Rosia CLI
# This file serves as a template. GoReleaser will generate the actual formula
# and publish it to the homebrew-rosia tap repository.
#
# Manual installation (if not using GoReleaser):
# 1. Update the version, url, and sha256 values
# 2. Copy to homebrew-rosia repository as Formula/rosia.rb
# 3. Users can install with: brew install raucheacho/rosia/rosia

class Rosia < Formula
  desc "Clean development dependencies and caches across multiple project types"
  homepage "https://github.com/raucheacho/rosia-cli"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/raucheacho/rosia-cli/releases/download/v0.1.0/rosia_0.1.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_ARM64"
    else
      url "https://github.com/raucheacho/rosia-cli/releases/download/v0.1.0/rosia_0.1.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/raucheacho/rosia-cli/releases/download/v0.1.0/rosia_0.1.0_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"
    else
      url "https://github.com/raucheacho/rosia-cli/releases/download/v0.1.0/rosia_0.1.0_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"
    end
  end

  def install
    bin.install "rosia"
    
    # Install shell completions (if available)
    if File.exist?("completions/bash")
      bash_completion.install "completions/bash/rosia"
    end
    if File.exist?("completions/zsh")
      zsh_completion.install "completions/zsh/_rosia"
    end
    if File.exist?("completions/fish")
      fish_completion.install "completions/fish/rosia.fish"
    end
    
    # Install default profiles
    if Dir.exist?("profiles")
      (prefix/"profiles").install Dir["profiles/*"]
    end
  end

  def post_install
    # Create configuration directory
    config_dir = "#{Dir.home}/.rosia"
    unless Dir.exist?(config_dir)
      mkdir_p config_dir
      mkdir_p "#{config_dir}/trash"
      mkdir_p "#{config_dir}/plugins"
      mkdir_p "#{config_dir}/profiles"
    end
    
    # Copy default profiles if they don't exist
    if Dir.exist?("#{prefix}/profiles")
      Dir["#{prefix}/profiles/*"].each do |profile|
        profile_name = File.basename(profile)
        dest = "#{config_dir}/profiles/#{profile_name}"
        cp profile, dest unless File.exist?(dest)
      end
    end
    
    # Create default config if it doesn't exist
    config_file = "#{Dir.home}/.rosiarc.json"
    unless File.exist?(config_file)
      default_config = {
        trash_retention_days: 3,
        profiles: ["node", "python", "rust", "flutter", "go"],
        ignore_paths: [],
        plugins: [],
        concurrency: 0,
        telemetry_enabled: false
      }
      File.write(config_file, JSON.pretty_generate(default_config))
    end
  end

  test do
    # Test that the binary runs and shows version
    assert_match version.to_s, shell_output("#{bin}/rosia version")
    
    # Test that help command works
    assert_match "Clean development dependencies", shell_output("#{bin}/rosia --help")
  end

  def caveats
    <<~EOS
      Rosia CLI has been installed!
      
      Configuration file: ~/.rosiarc.json
      Profiles directory: ~/.rosia/profiles
      Trash directory: ~/.rosia/trash
      
      Quick start:
        rosia scan ~/projects    # Scan for cleanable files
        rosia ui ~/projects      # Launch interactive TUI
        rosia clean              # Clean detected targets
        rosia stats              # View statistics
      
      Documentation: https://github.com/raucheacho/rosia-cli
    EOS
  end
end
