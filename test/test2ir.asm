# This test demonstrates register re-use when a variable is already loaded.

	.data
v1:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, 1		# v1 -> $t1
	# Register for v1 should be re-used at this point.
	li $t1, 4		# v1 -> $t1

	# Store variables back into memory
	sw $t1, v1
	li $v0, 10
	syscall
	.end main
