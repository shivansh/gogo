	.data
a.0:		.word	0
t0:		.word	0
t1:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	# Store dirty variables back into memory
	sw	$3, a.0
	bne	$3, 2, l2

	lw	$3, a.0		# a.0 -> $3
	addi	$5, $3, 1
	move	$3, $5		# a.0 -> $3
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, t0
	j	l0

l2:
	lw	$3, a.0		# a.0 -> $3
	addi	$5, $3, 4
	move	$3, $5		# a.0 -> $3
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, t1
	j	l0

l0:
	li	$2, 10
	syscall
	.end main
