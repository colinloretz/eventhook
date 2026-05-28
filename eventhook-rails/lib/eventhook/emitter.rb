module EventHook
  module Emitter
    # EventHook.emit('payment.completed', { order_id: 123 })
    # EventHook.emit('user.created', { id: 1 }, idempotency_key: "user-1", source: :internal)
    def self.emit(event_type, payload, idempotency_key: nil, source: nil, headers: {})
      client.post_event(
        event_type:      event_type,
        payload:         payload,
        idempotency_key: idempotency_key,
        headers:         headers
      )
    rescue Faraday::Error => e
      logger.error("[EventHook] emit failed: #{e.message}")
      nil
    end

    private_class_method def self.client
      @client ||= Client.new
    end

    private_class_method def self.logger
      EventHook.configuration.logger ||
        (defined?(Rails) ? Rails.logger : Logger.new($stdout))
    end
  end
end
