package database_test

import (
	. "github.com/cloudfoundry/mysql-diag/database"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
		It("is bootstrappable", func() {
			nodes := []*GaleraStatus{nodeSyncedReadOnly, nodeSyncedReadOnly, nodeSyncedReadOnly}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeTrue())
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
		It("is bootstrappable", func() {
			nodes := []*GaleraStatus{nodeJoinerWritable, nodeSyncedReadOnly, nodeSyncedReadOnly}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeTrue())
		})
	})
	Context("some nodes Joiner, some nodes Synced but readonly, some nodes are missing", func() {
		It("is bootstrappable", func() {
			nodes := []*GaleraStatus{nodeJoinerWritable, nodeSyncedReadOnly, nil}
			needed := DoWeNeedBootstrap(nodes)
			Expect(needed).To(BeTrue())
		})
	})
})
