require 'eventhook/version'
require 'eventhook/configuration'
require 'eventhook/client'
require 'eventhook/emitter'
require 'eventhook/engine' if defined?(Rails)

module EventHook
  class Error < StandardError; end
  class SignatureVerificationError < Error; end
  class ConfigurationError < Error; end

  class << self
    def configure
      yield configuration
    end

    def configuration
      @configuration ||= Configuration.new
    end

    # EventHook.emit('payment.completed', { order_id: 123 })
    def emit(event_type, payload, **opts)
      Emitter.emit(event_type, payload, **opts)
    end
  end
end
