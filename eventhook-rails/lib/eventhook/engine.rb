require 'rails/engine'

module EventHook
  class Engine < ::Rails::Engine
    isolate_namespace EventHook

    initializer 'eventhook.load_sources' do
      require 'eventhook/sources/base'
      require 'eventhook/sources/stripe'
      require 'eventhook/sources/github'
      require 'eventhook/sources/shopify'
      require 'eventhook/sources/generic'
    end
  end
end
