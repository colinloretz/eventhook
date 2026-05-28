require 'faraday'
require 'json'

module EventHook
  class Client
    def initialize(config = EventHook.configuration)
      @config = config
    end

    def post_event(event_type:, payload:, source_id: nil, idempotency_key: nil, headers: {})
      body = { event_type: event_type, payload: payload, headers: headers }
      body[:source_id]       = source_id       if source_id
      body[:idempotency_key] = idempotency_key if idempotency_key

      post('/api/v1/events', body)
    end

    def get(path, params = {})
      connection.get(path, params) do |req|
        req.headers['Authorization'] = "Bearer #{@config.api_key}"
      end.body
    end

    def post(path, body = {})
      connection.post(path) do |req|
        req.headers['Authorization'] = "Bearer #{@config.api_key}"
        req.headers['Content-Type']  = 'application/json'
        req.body = JSON.generate(body)
      end.body
    end

    private

    def connection
      @connection ||= Faraday.new(url: @config.runtime_url) do |f|
        f.options.timeout      = @config.timeout
        f.options.open_timeout = @config.timeout
        f.response :json
        f.adapter Faraday.default_adapter
      end
    end
  end
end
