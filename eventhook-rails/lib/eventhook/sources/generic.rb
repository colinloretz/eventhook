require 'openssl'

module EventHook
  module Sources
    # Generic HMAC-SHA256 verification with configurable header name.
    class Generic < Base
      def initialize(secret, header: 'X-Webhook-Signature', prefix: 'sha256=')
        super(secret)
        @header = header
        @prefix = prefix
      end

      def verify!(request)
        header_key = 'HTTP_' + @header.upcase.tr('-', '_')
        value = request.headers[header_key] ||
                request.headers[@header] ||
                raise(SignatureVerificationError, "Missing #{@header} header")

        payload  = request.raw_post
        digest   = OpenSSL::HMAC.hexdigest('SHA256', @secret, payload)
        expected = "#{@prefix}#{digest}"

        unless secure_compare(value, expected)
          raise SignatureVerificationError, 'Signature mismatch'
        end

        'webhook'
      end
    end
  end
end
