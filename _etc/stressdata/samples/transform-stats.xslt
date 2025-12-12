<?xml version="1.0"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
    <xsl:output method="text"/>

    <xsl:template match="/">
        StartOn is: <xsl:value-of select="/StressStats/StartOn"/>
        Target IP is <xsl:value-of select="/StressStats/Config/@targetIP"/>

        <xsl:apply-templates select="/StressStats/Simulators/Providers/Latches/Providers"/>
    </xsl:template>


    <xsl:template match="Providers">
        Provider Directory: <xsl:value-of select="./@directory"/>

        <xsl:for-each select="File/Stats">
            Sum: <xsl:value-of select="Iterations + SendCount"/>
        </xsl:for-each>
    </xsl:template>
</xsl:stylesheet>