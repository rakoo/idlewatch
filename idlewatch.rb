require 'net/imap'
require 'trollop'

opts = Trollop::options do
  opt :login, "Login", :type => :string
  opt :pass, "Pass", :type => :string
  opt :mailbox, "The mailbox to watch", :default => "[Gmail]/All Mail"
end

raise "No login !" unless opts[:login]
raise "No password !" unless opts[:pass]

def debug str
  puts "[#{Time.now.to_s}] #{str}"
end

debug "Starting idle loop over here"
loop do

  imap = Net::IMAP.new 'imap.gmail.com', ssl: true unless imap

  imap.login opts[:login], opts[:pass]
  imap.select opts[:mailbox]

  debug "Watching #{opts[:mailbox]}"

  Thread.new do
    debug "Starting timer"
    sleep 29 * 60
    imap.idle_done
  end

  Thread.new do
    debug "Starting ssl timer"
    sleep 10 * 3600 + 59 * 60
    imap.idle_done
  end

  begin
    imap.idle do |resp|
      if resp.kind_of?(Net::IMAP::ContinuationRequest) and resp.data.text == 'idling'
        debug "Starting idle loop over there"
      end

      if resp.kind_of?(Net::IMAP::UntaggedResponse) and resp.name == 'EXISTS'
        debug "Running sync" 
        system('offlineimap -u Quiet')
        debug "Ran sync"
      end

    end
  rescue Errno::ECONNRESET
    debug "! Connection reset by peer"
  rescue Net::IMAP::Error => error
    debug "! Imap error : #{error.inspect}"
  rescue Errno::ETIMEDOUT
    debug "! timeout !"
  rescue EOFError
    debug "! EOF"
  end

end

