require 'openssl'

module EventHook
  module Sources
    # Verifies X-Hub-Signature-256 header (GitHub webhooks).
    # https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
    class Github < Base
      def verify!(request)
        header  = request.headers['HTTP_X_HUB_SIGNATURE_256'] ||
                  request.headers['X-Hub-Signature-256'] ||
                  raise(SignatureVerificationError, 'Missing X-Hub-Signature-256 header')
        payload = request.raw_post

        expected = 'sha256=' + OpenSSL::HMAC.hexdigest('SHA256', @secret, payload)

        unless secure_compare(header, expected)
          raise SignatureVerificationError, 'GitHub signature mismatch'
        end

        event_type_from(request)
      end

      private

      def event_type_from(request)
        event = request.headers['HTTP_X_GITHUB_EVENT'] ||
                request.headers['X-GitHub-Event'] ||
                'github.webhook'
        "github.#{event}"
      end
    end
  end
end
