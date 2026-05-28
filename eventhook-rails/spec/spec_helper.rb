require 'openssl'
require 'base64'
require 'json'

# Load gem without Rails
$LOAD_PATH.unshift File.expand_path('../lib', __dir__)
require 'eventhook/version'
require 'eventhook/sources/base'
require 'eventhook/sources/stripe'
require 'eventhook/sources/github'
require 'eventhook/sources/shopify'
require 'eventhook/sources/generic'

module EventHook
  class SignatureVerificationError < StandardError; end
end

RSpec.configure do |config|
  config.expect_with :rspec do |c|
    c.syntax = :expect
  end
end
