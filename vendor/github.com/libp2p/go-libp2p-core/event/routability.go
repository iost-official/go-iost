package event

// EvtLocalRoutabilityPrivate is an event struct to be emitted with the local's
// node routability changes to PRIVATE (i.e. not routable from the Internet).
//
// This event is usually emitted by the AutoNAT subsystem.
type EvtLocalRoutabilityPrivate struct{}

// EvtLocalRoutabilityPublic is an event struct to be emitted with the local's
// node routability changes to PUBLIC (i.e. appear to routable from the
// Internet).
//
// This event is usually emitted by the AutoNAT subsystem.
type EvtLocalRoutabilityPublic struct{}

// EvtLocalRoutabilityUnknown is an event struct to be emitted with the local's
// node routability changes to UNKNOWN (i.e. we were unable to make a
// determination about our NAT status with enough confidence).
//
// This event is usually emitted by the AutoNAT subsystem.
type EvtLocalRoutabilityUnknown struct{}
