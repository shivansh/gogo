	.data
a.0:		.word	0
b.1:		.word	0
y.3:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	li	$5, 2		# b.1 -> $5
	sw	$5, b.1	# spilled b.1, freed $5
	move	$5, $3		# y.3 -> $5
	li	$2, 1
	move	$4, $5
	syscall
	li	$3, 4		# a.0 -> $3
	li	$2, 1
	move	$4, $3
	syscall
	move	$6, $3		# b.1 -> $6
	li	$2, 1
	move	$4, $6
	syscall
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, y.3
	sw	$6, b.1
	li	$2, 10
	syscall
	.end main
