require 'openssl'
require 'base64'

module EventHook
  module Sources
    # Verifies X-Shopify-Hmac-SHA256 header (Shopify webhooks).
    # https://shopify.dev/docs/apps/build/webhooks/secure/https-webhooks#verify-the-webhook
    class Shopify < Base
      def verify!(request)
        header  = request.headers['HTTP_X_SHOPIFY_HMAC_SHA256'] ||
                  request.headers['X-Shopify-Hmac-SHA256'] ||
                  raise(SignatureVerificationError, 'Missing X-Shopify-Hmac-SHA256 header')
        payload = request.raw_post

        expected = Base64.strict_encode64(OpenSSL::HMAC.digest('SHA256', @secret, payload))

        unless secure_compare(header, expected)
          raise SignatureVerificationError, 'Shopify signature mismatch'
        end

        event_type_from(request)
      end

      private

      def event_type_from(request)
        topic = request.headers['HTTP_X_SHOPIFY_TOPIC'] ||
                request.headers['X-Shopify-Topic'] ||
                'webhook'
        "shopify.#{topic.tr('/', '.')}"
      end
    end
  end
end
