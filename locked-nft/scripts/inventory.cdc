import NFTLocker from "../contracts/NFTLocker.cdc"
import ExampleNFT from 0xEXAMPLENFTADDRESS

pub fun main(acctAddress: Address): [UInt64]? {
    let nftOwner = getAccount(acctAddress);
    let capability = nftOwner.getCapability<&{NFTLocker.LockedCollection}>(NFTLocker.CollectionPublicPath);
    let borrowed = capability.borrow() ?? panic("Could not borrow receiver reference")
    return borrowed.getIDs(nftType: Type<@ExampleNFT.NFT>())
}