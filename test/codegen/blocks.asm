	.data
withinBlock.0:	.asciiz "\nInside the block\nValue of a: "
outsideBlock.1:	.asciiz "\nOutside the block\nValue of a:"
a.2:		.word	0
a.3:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 1		# a.2 -> $3
	li	$2, 4
	la	$4, outsideBlock.1
	syscall
	li	$2, 1
	move	$4, $3
	syscall
	li	$5, 4		# a.3 -> $5
	li	$2, 4
	la	$4, withinBlock.0
	syscall
	li	$2, 1
	move	$4, $5
	syscall
	li	$2, 4
	la	$4, outsideBlock.1
	syscall
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, a.2
	sw	$5, a.3
	li	$2, 10
	syscall
	.end main
