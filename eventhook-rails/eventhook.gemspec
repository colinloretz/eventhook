require_relative 'lib/eventhook/version'

Gem::Specification.new do |s|
  s.name        = 'eventhook'
  s.version     = EventHook::VERSION
  s.summary     = 'Stripe-quality webhook observability for every event in your app'
  s.description = 'Drop-in webhook infrastructure for Rails. Inbound verification, outbound delivery, full dashboard.'
  s.homepage    = 'https://eventhook.dev'
  s.license     = 'MIT'
  s.authors     = ['EventHook']

  s.files = Dir['lib/**/*', 'app/**/*', 'config/**/*', 'README.md']
  s.require_paths = ['lib']

  s.required_ruby_version = '>= 3.0'

  s.add_dependency 'railties', '>= 6.0'
  s.add_dependency 'faraday', '>= 2.0'

  s.add_development_dependency 'rspec-rails'
  s.add_development_dependency 'webmock'
  s.add_development_dependency 'rails', '>= 6.0'
end
