	.data
a.x.0:		.word	0
a.y.1:		.word	0
b.x.2:		.word	0
b.y.3:		.word	0
c.4:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 0		# a.x.0 -> $3
	sw	$3, a.x.0	# spilled a.x.0, freed $3
	li	$3, 0		# a.y.1 -> $3
	sw	$3, a.y.1	# spilled a.y.1, freed $3
	li	$3, 3		# b.x.2 -> $3
	li	$5, 3		# b.y.3 -> $5
	sw	$5, b.y.3	# spilled b.y.3, freed $5
	move	$5, $3		# c.4 -> $5
	# Store dirty variables back into memory
	sw	$3, b.x.2
	sw	$5, c.4
	li	$2, 10
	syscall
	.end main
