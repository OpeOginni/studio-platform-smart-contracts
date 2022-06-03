import Golazo from "../../../contracts/Golazo.cdc"

transaction(name: String) {
    // local variable for the admin reference
    let admin: &Golazo.Admin

    prepare(signer: AuthAccount) {
        // borrow a reference to the Admin resource
        self.admin = signer.borrow<&Golazo.Admin>(from: Golazo.AdminStoragePath)
            ?? panic("Could not borrow a reference to the Golazo Admin capability")
    }

    execute {
        let id = self.admin.createSet(
            name: name,
        )

        log("====================================")
        log("New Set: ".concat(name))
        log("SetID: ".concat(id.toString()))
        log("====================================")
    }
}

