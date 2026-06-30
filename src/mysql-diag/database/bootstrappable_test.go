package database_test

import (
	"crypto/tls"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/mysql-diag/database"
)

var _ = Describe("IsTLSError", func() {
	It("returns true for a tls.CertificateVerificationError", func() {
		err := &tls.CertificateVerificationError{Err: fmt.Errorf("x509: certificate is valid for localhost, not mysql-diag")}
		Expect(IsTLSError(err)).To(BeTrue())
	})

	It("returns true when the TLS error is wrapped", func() {
		inner := &tls.CertificateVerificationError{Err: fmt.Errorf("hostname mismatch")}
		wrapped := fmt.Errorf("dial failed: %w", inner)
		Expect(IsTLSError(wrapped)).To(BeTrue())
	})

	It("returns false for a plain connection error", func() {
		Expect(IsTLSError(fmt.Errorf("connection refused"))).To(BeFalse())
	})

	It("returns false for nil", func() {
		Expect(IsTLSError(nil)).To(BeFalse())
	})
})

var _ = Describe("HasTLSErrors", func() {
	tlsErr := &tls.CertificateVerificationError{Err: fmt.Errorf("x509: certificate is valid for localhost, not mysql-diag")}

	It("returns true when any node has a TLS error", func() {
		rows := map[string]*NodeClusterStatus{
			"10.0.0.1": {Err: tlsErr},
			"10.0.0.2": {Err: tlsErr},
			"10.0.0.3": {Err: tlsErr},
		}
		Expect(HasTLSErrors(rows)).To(BeTrue())
	})

	It("returns true when only some nodes have TLS errors", func() {
		rows := map[string]*NodeClusterStatus{
			"10.0.0.1": {Err: tlsErr},
			"10.0.0.2": {Status: &GaleraStatus{LocalState: "Synced"}},
		}
		Expect(HasTLSErrors(rows)).To(BeTrue())
	})

	It("returns false when errors are non-TLS", func() {
		rows := map[string]*NodeClusterStatus{
			"10.0.0.1": {Err: fmt.Errorf("connection refused")},
		}
		Expect(HasTLSErrors(rows)).To(BeFalse())
	})

	It("returns false when there are no errors", func() {
		rows := map[string]*NodeClusterStatus{
			"10.0.0.1": {Status: &GaleraStatus{LocalState: "Synced"}},
		}
		Expect(HasTLSErrors(rows)).To(BeFalse())
	})
})

var _ = Describe("bootstrappable", func() {

	nodeSyncedWritable := &GaleraStatus{
		LocalState: "Synced",
		ReadOnly:   false,
	}
	nodeSyncedReadOnly := &GaleraStatus{
		LocalState: "Synced",
		ReadOnly:   true,
	}
	nodeDonorDesyncedWritable := &GaleraStatus{
		LocalState: "Donor/Desynced",
		ReadOnly:   false,
	}
	nodeJoinerWritable := &GaleraStatus{
		LocalState: "Joiner",
		ReadOnly:   false,
	}

	Context("at least one writable Synced", func() {
		It("is not bootstrappable", func() {
			nodes := []*GaleraStatus{nodeSyncedReadOnly, nodeSyncedReadOnly, nodeSyncedWritable}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeFalse())
		})
	})
	Context("at least one writable Donor/Desynced", func() {
		It("is not bootstrappable", func() {
			nodes := []*GaleraStatus{nodeSyncedReadOnly, nodeSyncedReadOnly, nodeDonorDesyncedWritable}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeFalse())
		})
	})
	Context("all nodes Joiner", func() {
		It("is bootstrappable", func() {
			nodes := []*GaleraStatus{nodeJoinerWritable, nodeJoinerWritable, nodeJoinerWritable}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeTrue())
		})
	})
	Context("all nodes Synced but readonly", func() {
		It("does not need bootstrap with any Synced nodes", func() {
			nodes := []*GaleraStatus{nodeSyncedReadOnly, nodeSyncedReadOnly, nodeSyncedReadOnly}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeFalse())
		})
	})
	Context("all nodes missing", func() {
		It("is bootstrappable", func() {
			nodes := []*GaleraStatus{nil, nil, nil}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeTrue())
		})
	})
	Context("some nodes Joiner, some nodes Synced but readonly", func() {
		It("does not need bootstrap with any Synced nodes", func() {
			nodes := []*GaleraStatus{nodeJoinerWritable, nodeSyncedReadOnly, nodeSyncedReadOnly}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeFalse())
		})
	})
	Context("some nodes Joiner, some nodes Synced but readonly, some nodes are missing", func() {
		It("does not need bootstrap with any Synced nodes", func() {
			nodes := []*GaleraStatus{nodeJoinerWritable, nodeSyncedReadOnly, nil}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeFalse())
		})
	})
})
