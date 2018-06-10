	.data
a.0:		.word	0
c.1:		.word	0
b.2:		.word	0

	.text


	.globl main
	.ent main
main:
	li	$3, 2		# a.0 -> $3
	sw	t3, a.0		# global decl -> memory
	li	$3, 8		# b.2 -> $3
	sw	t3, b.2		# global decl -> memory
	li	$3, 4		# c.1 -> $3
	# Store dirty variables back into memory
	sw	$3, c.1
	li	$2, 10
	syscall
	.end main
