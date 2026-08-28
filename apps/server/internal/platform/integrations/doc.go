// Package integrations defines provider metadata and capability discovery for
// first-party integration adapters compiled into FortyOne.
//
// It intentionally does not define a universal provider interface. Each use
// case owns a narrow capability port; the registry answers whether a provider
// declares a compatible version before bootstrap resolves that typed factory.
package integrations
