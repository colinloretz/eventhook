module EventHook
  class Configuration
    attr_accessor :runtime_url, :api_key, :environment, :timeout, :logger

    def initialize
      @runtime_url = ENV.fetch('EVENTHOOK_URL', 'http://localhost:7676')
      @api_key     = ENV.fetch('EVENTHOOK_KEY', 'dev-api-key')
      @environment = ENV.fetch('RAILS_ENV', 'development')
      @timeout     = 5
      @sources     = {}
    end

    # config.sources do |s|
    #   s.add :stripe,  secret: ENV['STRIPE_WEBHOOK_SECRET']
    #   s.add :github,  secret: ENV['GITHUB_WEBHOOK_SECRET']
    # end
    def sources
      yield SourceRegistry.new(@sources) if block_given?
      @sources
    end

    def source(slug)
      @sources[slug.to_sym]
    end
  end

  class SourceRegistry
    def initialize(store)
      @store = store
    end

    def add(slug, secret:, type: nil)
      @store[slug.to_sym] = { secret: secret, type: (type || slug).to_sym }
    end
  end
end
