# This test demonstrates register re-use when a variable is already loaded.

	.data
This test demonstrates register re-use when a variable is already loaded.:	.word	0
v1:	.word	0
Register for v1 should be re-used at this point.:	.word	0

	.text
	.globl main
	.ent main
main:
	li $t1, 1		# v1 -> $t1
	# Register for v1 should be re-used at this point.
	li $t1, 4		# v1 -> $t1

	# Store variables back into memory
	sw $t1, v1
