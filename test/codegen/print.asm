	.data
a.0:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 4		# a.0 -> $3
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, a.0
	li	$2, 10
	syscall
	.end main
