EventHook::Engine.routes.draw do
  # Inbound webhook receiver — POST /eventhook/in/:source
  post 'in/:source', to: 'inbound#receive'

  # Dashboard + API proxy — GET/POST /eventhook/dashboard and /eventhook/api/*
  get  'dashboard',      to: 'dashboard#index'
  get  'dashboard/*path', to: 'dashboard#index'
  match 'api/*path',     to: 'dashboard#proxy', via: :all
end
