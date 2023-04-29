import NonFungibleToken from "./NonFungibleToken.cdc"

pub contract NFTLocker {

    pub event Withdraw(id: UInt64, from: Address?)
    pub event Deposit(id: UInt64, to: Address?)
    pub event NFTLocked(
        id: UInt64,
        to: Address?,
        lockedAt: UInt64,
        lockedUntil: UInt64,
        duration: UInt64,
        nftType: Type
    )
    pub event NFTUnlocked(
        id: UInt64,
        from: Address?,
        nftType: Type
    )

    pub let CollectionStoragePath:  StoragePath
    pub let CollectionPublicPath:   PublicPath

    pub var totalLockedTokens:      UInt64
    access(self) let lockedTokens:  {Type: {UInt64: LockedData}}

    pub struct LockedData {
        pub let id: UInt64
        pub let owner: Address
        pub let lockedAt: UInt64
        pub let lockedUntil: UInt64
        pub let duration: UInt64
        pub let nftType: Type

        init (id: UInt64, owner: Address, duration: UInt64, nftType: Type) {
            if let lockedToken = (NFTLocker.lockedTokens[nftType]!)[id] {
                self.id = id
                self.owner = lockedToken.owner
                self.lockedAt = lockedToken.lockedAt
                self.lockedUntil = lockedToken.lockedUntil
                self.duration = lockedToken.duration
                self.nftType = lockedToken.nftType
            } else {
                self.id = id
                self.owner = owner
                self.lockedAt = UInt64(getCurrentBlock().timestamp)
                self.lockedUntil = self.lockedAt + duration
                self.duration = duration
                self.nftType = nftType
            }
        }
    }

    pub fun getNFTLockerDetails(id: UInt64, nftType: Type): NFTLocker.LockedData? {
        return (NFTLocker.lockedTokens[nftType]!)[id]!
    }

    pub fun canUnlockToken(id: UInt64, nftType: Type): Bool {
        if let lockedToken = (NFTLocker.lockedTokens[nftType]!)[id] {
            if lockedToken.lockedUntil < UInt64(getCurrentBlock().timestamp) {
                return true
            }
        }

        return false
    }

    pub resource interface LockedCollection {
        pub fun getIDs(nftType: Type): [UInt64]?
    }

    pub resource interface LockProvider {
        pub fun lock(token: @NonFungibleToken.NFT, duration: UInt64)
        pub fun unlock(id: UInt64, nftType: Type): @NonFungibleToken.NFT
    }

    pub resource Collection: LockedCollection, LockProvider {
        pub var lockedNFTs: @{Type: {UInt64: NonFungibleToken.NFT}}

        pub fun unlock(id: UInt64, nftType: Type): @NonFungibleToken.NFT {
            pre {
                NFTLocker.canUnlockToken(
                    id: id,
                    nftType: nftType
                ) == true : "locked duration has not been met"
            }

            let token <- self.lockedNFTs[nftType]?.remove(key: id)!!

            if let lockedToken = NFTLocker.lockedTokens[nftType] {
                lockedToken.remove(key: id)
            }
            NFTLocker.totalLockedTokens = NFTLocker.totalLockedTokens - 1

            emit NFTUnlocked(
                id: token.id,
                from: self.owner?.address,
                nftType: nftType
            )

            emit Withdraw(id: token.id, from: self.owner?.address)

            return <-token
        }

        pub fun lock(token: @NonFungibleToken.NFT, duration: UInt64) {
            let id: UInt64 = token.id
            let nftType: Type = token.getType()

            if NFTLocker.lockedTokens[nftType] == nil {
                NFTLocker.lockedTokens[nftType] = {}
            }

            if self.lockedNFTs[nftType] == nil {
                self.lockedNFTs[nftType] <-! {}
            }

            let oldToken <- self.lockedNFTs.insert(key: nftType, <-{id: <- token})

            let nestedLock = NFTLocker.lockedTokens[nftType]!
            let lockedData = NFTLocker.LockedData(
                id: id,
                owner: self.owner!.address,
                duration: duration,
                nftType: nftType
            )
            nestedLock[id] = lockedData
            NFTLocker.lockedTokens[nftType] = nestedLock

            NFTLocker.totalLockedTokens = NFTLocker.totalLockedTokens + 1

            emit NFTLocked(
                id: id,
                to: self.owner?.address,
                lockedAt: lockedData.lockedAt,
                lockedUntil: lockedData.lockedUntil,
                duration: lockedData.duration,
                nftType: nftType
            )

            emit Deposit(id: id, to: self.owner?.address)

            destroy oldToken
        }

        pub fun getIDs(nftType: Type): [UInt64]? {
            return self.lockedNFTs[nftType]?.keys
        }

        destroy() {
            destroy self.lockedNFTs
        }

        init() {
            self.lockedNFTs <- {}
        }
    }

    pub fun createEmptyCollection(): @Collection {
        return <- create Collection()
    }

    init() {
        self.CollectionStoragePath = /storage/NFTLockerCollection
        self.CollectionPublicPath = /public/NFTLockerCollection

        self.totalLockedTokens = 0
        self.lockedTokens = {}
    }
}