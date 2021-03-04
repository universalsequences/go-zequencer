package main

const SampleCreated = "SampleCreated(address,bytes32,string,uint32)"
const SampleTagged = "SampleTagged(bytes32,bytes32,uint32)"
const SampleYoutube = "SampledYoutube(bytes32,bytes32,uint32)"
const SampleYear = "SampleYear(bytes32,int16,uint32)"
const NewDiscogsSample = "NewDiscogsSample(bytes32,uint256,bytes32,uint32)"
const ReleaseInfo = "ReleaseInfo(bytes32,uint32,string,bytes32,bytes32)"
const Sample = "ReleaseInfo(bytes32,uint32,string,bytes32,bytes32)"
const SequenceEdited = "SequenceEdited(address,bytes32,bytes32,string)"
const SampleInSequence = "SampleInSequence(bytes32,bytes32)"
const SampleByArtist = "SampleByArtist(bytes32,bytes32)"
const NewPreset = "NewPreset(address,bytes32,uint32,bytes32,bytes32[9],bytes32[6])"
const PresetInstrumentType = "PresetInstrumentType(bytes32,bytes32,uint32)"
const PresetTagged = "PresetTagged(address,bytes32,bytes32,uint32)"
const GuildKeyEncrypted = "GuildKeyEncrypted(address,uint32,bytes32)"
const NewEncryptedContent = "NewEncryptedContent(address,bytes32,bytes32,int8,bytes32,bytes32[9],bytes32[6])"
const NewGuildMemberRequestAccepted  = "NewGuildMemberRequestAccepted(address,uint32)"
const EncryptedContentShared = "EncryptedContentShared(address,address,bytes32,bytes32[9],bytes32[6])"
const NewGuildSequence = "NewGuildSequence(address,bytes32,uint32,bytes32,bytes32,bytes32[9],bytes32[6])"
const SequenceStarred = "SequenceStarred(bytes32,address,uint32)"
const SequenceUnstarred = "SequenceUnstarred(bytes32,address,uint32)"
const SequenceFavorited = "SequenceFavorited(bytes32,address,uint32)"
const SequenceUnfavorited = "SequenceUnfavorited(bytes32,address,uint32)"
const NewAnnotation = "NewAnnotation(bytes32,bytes32,bytes32,address)"
const NewPack = "NewPack(address,bytes32,bytes32,bytes32)"
    
// have max 2 keys to index on - so that they can do sub-sorting based on the other key
var TableIndices = map[string][]string {
	SampleCreated: []string{"guildId", "ipfsHash"},
	SampleTagged: []string{"tag", "guildId", "ipfsHash"},
	SampleYear: []string{"year", "ipfsHash"},
	SampleYoutube: []string{"videoId", "ipfsHash"},
	NewDiscogsSample: []string{"sampleHash", "discogsId"},
	ReleaseInfo: []string{"artistName", "releaseId"},
	SampleByArtist: []string{"artistName", "ipfsHash"},
	SampleInSequence: []string{"sequenceHash", "sampleHash"},
	SequenceEdited: []string{"newSequence", "previousSequence"},
	NewPreset: []string{"user", "contentHash"},
	PresetInstrumentType: []string{"contentHash","instrumentType"},
	PresetTagged: []string{"tag","contentHash"},
	GuildKeyEncrypted: []string{"guildId","user"},
	NewEncryptedContent: []string{"newContent","blockNumber"},
	NewGuildMemberRequestAccepted: []string{"newMember", "guildId"},
	EncryptedContentShared: []string{"contentHash", "sharedWith"},
	NewGuildSequence: []string{"contentHash", "guildId"},
	SequenceStarred: []string{"ipfsHash", "guildId"},
	SequenceUnstarred: []string{"ipfsHash",  "guildId"},
	SequenceFavorited: []string{"ipfsHash", "user", "guildId"},
	SequenceUnfavorited: []string{"ipfsHash", "user", "guildId"},
	NewAnnotation: []string{"data", "annotationType"},
	NewPack: []string{"packHash", "user"},
}
