require 'net/http'
require 'uri'

module EventHook
  # Reverse-proxies all /eventhook/dashboard/* and /eventhook/api/* requests
  # to the EventHook runtime. This means the React SPA is served from a single
  # binary and the Rails app never needs to know about the runtime's internals.
  class DashboardController < ApplicationController
    def index
      proxy_to_runtime(build_runtime_path)
    end

    def proxy
      proxy_to_runtime(build_runtime_path)
    end

    private

    def build_runtime_path
      # Map /eventhook/dashboard/* → /dashboard/*
      # Map /eventhook/api/*       → /api/*
      request.fullpath.sub(%r{\A/[^/]+}, '')
    end

    def proxy_to_runtime(path)
      runtime_url = EventHook.configuration.runtime_url
      uri         = URI.parse("#{runtime_url}#{path}")
      uri.query   = request.query_string.presence

      http = Net::HTTP.new(uri.host, uri.port)
      http.open_timeout = 5
      http.read_timeout = 30

      upstream_request = build_upstream_request(uri)

      upstream_response = http.request(upstream_request)

      # Forward status, selected headers, and body
      response.status = upstream_response.code.to_i
      %w[content-type cache-control etag last-modified].each do |h|
        response.set_header(h, upstream_response[h]) if upstream_response[h]
      end
      render plain: upstream_response.body, content_type: upstream_response['content-type'] || 'text/html'
    rescue => e
      logger.error("[EventHook] Dashboard proxy error: #{e.message}")
      render plain: 'EventHook runtime unavailable', status: :bad_gateway
    end

    def build_upstream_request(uri)
      klass = {
        'GET'    => Net::HTTP::Get,
        'POST'   => Net::HTTP::Post,
        'PUT'    => Net::HTTP::Put,
        'PATCH'  => Net::HTTP::Patch,
        'DELETE' => Net::HTTP::Delete,
      }.fetch(request.method, Net::HTTP::Get)

      req = klass.new(uri)
      req['Authorization'] = "Bearer #{EventHook.configuration.api_key}"
      req['Content-Type']  = request.content_type if request.content_type
      req['Accept']        = request.accept       if request.accept

      if request.body.present?
        req.body = request.body.read
        request.body.rewind
      end

      req
    end
  end
end
