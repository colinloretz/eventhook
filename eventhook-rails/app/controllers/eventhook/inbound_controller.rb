require 'eventhook/sources/base'
require 'eventhook/sources/stripe'
require 'eventhook/sources/github'
require 'eventhook/sources/shopify'
require 'eventhook/sources/generic'

module EventHook
  class InboundController < ApplicationController
    # POST /eventhook/in/:source
    #
    # 1. Look up source config by slug
    # 2. Verify signature using source-specific strategy
    # 3. Forward verified payload to the runtime
    # 4. Return 200 immediately (async delivery handled by runtime)
    def receive
      source_slug = params[:source]
      source_cfg  = EventHook.configuration.source(source_slug)

      unless source_cfg
        render json: { error: 'unknown source' }, status: :not_found
        return
      end

      verifier   = build_verifier(source_cfg)
      event_type = verifier.verify!(request)

      payload = parse_payload
      client  = Client.new

      client.post_event(
        event_type: event_type,
        payload:    payload,
        headers:    extract_headers
      )

      render json: { status: 'accepted' }, status: :ok
    rescue SignatureVerificationError => e
      logger.warn("[EventHook] Signature verification failed for '#{params[:source]}': #{e.message}")
      render json: { error: 'signature verification failed' }, status: :unauthorized
    rescue => e
      logger.error("[EventHook] Inbound error for '#{params[:source]}': #{e.message}")
      render json: { error: 'internal error' }, status: :internal_server_error
    end

    private

    def build_verifier(source_cfg)
      secret = source_cfg[:secret]
      case source_cfg[:type]
      when :stripe   then Sources::Stripe.new(secret)
      when :github   then Sources::Github.new(secret)
      when :shopify  then Sources::Shopify.new(secret)
      else                Sources::Generic.new(secret)
      end
    end

    def parse_payload
      body = request.raw_post
      JSON.parse(body)
    rescue JSON::ParserError
      { raw: body }
    end

    def extract_headers
      request.headers.each_with_object({}) do |(k, v), h|
        next unless k.start_with?('HTTP_') || %w[CONTENT_TYPE CONTENT_LENGTH].include?(k)
        h[k] = v
      end
    end
  end
end
