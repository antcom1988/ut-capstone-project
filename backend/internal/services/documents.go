package services

// RequiredDocumentPredicateSQL uses the fixed master-document alias m and
// employee alias p. ID 9 is required exactly for Fungsional, regardless of its
// global optional flag. All other requirements retain the explicit PRD/master
// flags (including SK CPNS and SK Pangkat for PPPK); no legal exemption is assumed.
// Reuse this predicate in both numerator and denominator for every projection.
const RequiredDocumentPredicateSQL = `(CASE WHEN m.id = 9 THEN p.jenis_jabatan = 'Fungsional' ELSE m.is_required END)`
