package cmd

import . "gopkg.in/check.v1"

func (s *TestSuite) TestConfigSections(c *C) {
	c.Assert(configSections, NotNil)
	c.Assert(len(configSections), Equals, 1)
}

func (s *TestSuite) TestGlobalConfigSection(c *C) {
	c.Assert(configSections["global"]("test"), Equals, "global.test")
}
