	.data
a.0:		.word	0
t0:		.word	0
b.1:		.word	0

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
	li	$3, -1		# a.0 -> $3
	addi	$5, $3, 1
	move	$6, $5		# b.1 -> $6
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, t0
	sw	$6, b.1
	li	$2, 10
	syscall
	.end main
