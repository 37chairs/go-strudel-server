#!/usr/bin/env ruby
# frozen_string_literal: true

require 'websocket-client-simple'
require 'json'
require 'time'
require 'securerandom'

class WebSocketClient
  def initialize
    @uri = 'ws://localhost:8069/ws'
    @client_id = SecureRandom.uuid
    @connected = false
  end

  def connect
    @ws = WebSocket::Client::Simple.connect @uri
    setup_event_handlers
    @connected = true
    puts "Connected to #{@uri}"
    true
  rescue StandardError => e
    puts "Failed to connect: #{e.message}"
    false
  end

  def send_message(type, content)
    return unless @connected && @ws

    message = {
      type: type,
      content: content,
      from: @client_id
    }

    begin
      @ws.send(JSON.generate(message))
      puts "Sent: #{type} - #{content}"
    rescue StandardError => e
      puts "Failed to send message: #{e.message}"
    end
  end

  def disconnect
    return unless @ws

    @ws.close
    @connected = false
    puts 'Disconnected from server'
  end

  private

  def setup_event_handlers
    @ws.on :open do
      puts 'WebSocket connection opened'
    end

    @ws.on :error do |e|
      puts "WebSocket error: #{e.message}"
    end

    @ws.on :close do |e|
      puts "WebSocket connection closed: #{e.code} #{e.reason}"
      @connected = false
    end
  end
end

def client
  client = WebSocketClient.new

  unless client.connect
    puts 'Failed to connect to server. Make sure the server is running.'
    return
  end

  puts "\nRuby WebSocket Client Commands:"
  puts '  setcps <cps> - Set the tempo' # setcps(1) is 120bpm I think
  puts '  play <pattern> - Send a play command'
  puts '  stop - Send a stop command'
  puts '  quit - Exit the client'
  puts ''

  while true
    print '> '
    input = gets&.strip

    break if input.nil? || input.downcase == 'quit'
    next if input.empty?

    parts = input.split(' ', 2)
    command = parts[0].downcase

    case command
    when 'setcps'
      if parts[1]
        client.send_message('setcps', {
                              'cps' => parts[1].to_f
                            })
      else
        puts 'Usage: setcps <cps>'
      end
    when 'play'
      if parts[1]
        client.send_message('play', {
                              'pattern' => parts[1]
                            })
      else
        puts 'Usage: play <pattern>'
      end
    when 'stop'
      client.send_message('stop', {})
    else
      puts "Unknown command. Type 'quit' to exit."
    end
  end

  client.disconnect
end

if __FILE__ == $PROGRAM_NAME
  puts 'Starting interactive Ruby WebSocket client...'
  client
end
