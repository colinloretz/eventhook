module EventHook
  module Sources
    class Base
      def initialize(secret)
        @secret = secret
      end

      # Returns the verified event_type string, or raises SignatureVerificationError.
      def verify!(request)
        raise NotImplementedError
      end

      protected

      def secure_compare(a, b)
        return false unless a.bytesize == b.bytesize
        ActiveSupport::SecurityUtils.secure_compare(a, b)
      rescue NameError
        # Fallback if ActiveSupport not available
        a.bytes.zip(b.bytes).reduce(0) { |acc, (x, y)| acc | (x ^ y) } == 0
      end
    end
  end
end
