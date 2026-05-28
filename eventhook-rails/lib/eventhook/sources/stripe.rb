require 'openssl'

module EventHook
  module Sources
    # Verifies Stripe-Signature header using Stripe's v1 HMAC scheme.
    # https://docs.stripe.com/webhooks#verify-manually
    class Stripe < Base
      TOLERANCE = 300 # seconds

      def verify!(request)
        header    = request.headers['HTTP_STRIPE_SIGNATURE'] ||
                    request.headers['Stripe-Signature'] ||
                    raise(SignatureVerificationError, 'Missing Stripe-Signature header')
        payload   = request.raw_post
        timestamp, signatures = parse_header(header)

        raise SignatureVerificationError, 'Request timestamp too old' if stale?(timestamp)

        signed_payload = "#{timestamp}.#{payload}"
        expected = OpenSSL::HMAC.hexdigest('SHA256', @secret, signed_payload)

        unless signatures.any? { |sig| secure_compare(sig, expected) }
          raise SignatureVerificationError, 'Stripe signature mismatch'
        end

        event_type_from(request)
      end

      private

      def parse_header(header)
        params = header.split(',').each_with_object({}) do |pair, h|
          k, v = pair.split('=', 2)
          h[k] ||= []
          h[k] << v
        end
        timestamp  = params.fetch('t', [nil]).first.to_i
        signatures = params.fetch('v1', [])
        [timestamp, signatures]
      end

      def stale?(timestamp)
        (Time.now.to_i - timestamp).abs > TOLERANCE
      end

      def event_type_from(request)
        body = JSON.parse(request.raw_post) rescue {}
        body['type'] || 'stripe.webhook'
      end
    end
  end
end
